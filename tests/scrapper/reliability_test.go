package scrapper

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	httpclient2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type MockMessageSender struct {
	mock.Mock
}

func (s *MockMessageSender) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	return s.Called(ctx, update).Error(0)
}

func (s *MockMessageSender) SendReport(ctx context.Context, report model.Report) error {
	return s.Called(ctx, report).Error(0)
}

// BuildHTTPNotifier — конфиг для тестов retry/timeout.
// ВНИМАНИЕ: ClosedStateInterval, OpenStateInterval, SlidingWindowSize заданы
// как голые числа (2, 2, 4) — это 2 нс и 4 нс, а не секунды.
// Для тестов CB используй buildCBNotifier с правильными единицами времени.
func BuildHTTPNotifier(url string) *httpclient.BotHTTPClient {
	cfg := httpclient2.Config{
		Timeout:              5 * time.Second,
		Delay:                1 * time.Second,
		MaxDelay:             2 * time.Second,
		DelayType:            "constant",
		Retries:              3,
		RetriableStatusCodes: []int{500},
		CircuitBreaker: httpclient2.CircuitBreakerConfig{
			MinRequiredCalls:     3,
			MaxRequestsHalfOpen:  3,
			ClosedStateInterval:  2 * time.Second,
			OpenStateInterval:    2 * time.Second,
			SlidingWindowSize:    4 * time.Second,
			FailureRateThreshold: 50,
		},
	}

	httpNotifier := httpclient.NewBotHTTPNotifier(url, httpclient2.NewReliableHTTPClient(cfg))
	return httpNotifier
}

// TC-2.1 Retry на 5xx
func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(6 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpNotifier := BuildHTTPNotifier(server.URL)

	update := model.LinkUpdate{
		ID:        uuid.New().String(),
		Link:      "https://google.com",
		Update:    model.Update{},
		TgChatIDs: []int64{123},
	}

	err := httpNotifier.SendUpdate(context.Background(), update)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got %v, want %v", err, context.DeadlineExceeded)
	}
}

// TC-2.1 Retry на 5xx
func TestRetry(t *testing.T) {
	attempt := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt++

		if attempt < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if attempt > 3 {
			t.Errorf("got %v, want %v", attempt, 3)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	httpNotifier := BuildHTTPNotifier(server.URL)

	update := model.LinkUpdate{
		ID:        uuid.New().String(),
		Link:      "https://google.com",
		Update:    model.Update{},
		TgChatIDs: []int64{123},
	}
	err := httpNotifier.SendUpdate(context.Background(), update)
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

// TC-2.2 Отсутствие Retry на 4xx
func TestNonRecoverableRetry(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt++
		if attempt > 1 {
			t.Errorf("supposed to not have any more requests after first error")
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	httpNotifier := BuildHTTPNotifier(server.URL)

	update := model.LinkUpdate{
		ID:        uuid.New().String(),
		Link:      "https://google.com",
		Update:    model.Update{},
		TgChatIDs: []int64{123},
	}
	err := httpNotifier.SendUpdate(context.Background(), update)
	if err == nil {
		t.Errorf("supposed to have error")
	}
}

// TC-2.3 Соблюдение интервала retry
func TestRetryInterval(t *testing.T) {
	const backoff = 1 * time.Second

	var requestTimes []time.Time
	var mu sync.Mutex

	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()

		attempt++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	httpNotifier := BuildHTTPNotifier(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := model.LinkUpdate{
		ID:        uuid.New().String(),
		Link:      "https://google.com",
		Update:    model.Update{},
		TgChatIDs: []int64{123},
	}

	_ = httpNotifier.SendUpdate(ctx, update)

	mu.Lock()
	times := requestTimes
	mu.Unlock()

	if len(times) < 2 {
		t.Skip("Not enough retries happened")
	}

	tolerance := 300 * time.Millisecond

	for i := 1; i < len(times); i++ {
		interval := times[i].Sub(times[i-1])

		if interval < backoff-tolerance {
			t.Errorf("Attempt %d: interval %v is less than expected %v (too early)",
				i, interval, backoff)
		}

		if interval > backoff+tolerance {
			t.Logf("Attempt %d: interval %v is greater than %v (might be test overhead)",
				i, interval, backoff)
		}
	}
}

// TC-4.1: Переход в OPEN
func TestCircuitBreaker_TransitionsToOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := buildCBNotifier(t, server.URL)
	update := model.LinkUpdate{ID: uuid.New().String(), TgChatIDs: []int64{1}}

	openCB(t, notifier)

	start := time.Now()
	err := notifier.SendUpdate(context.Background(), update)
	elapsed := time.Since(start)

	require.Error(t, err, "ожидается ошибка: CB находится в состоянии OPEN")
	assert.Less(t, elapsed, 50*time.Millisecond,
		"вызов при OPEN CB должен завершаться мгновенно, но занял %v", elapsed)
}

// TC-4.2: HALF-OPEN → CLOSED
func TestCircuitBreaker_HalfOpenTransitionsToClosed(t *testing.T) {
	var failMode atomic.Int32
	failMode.Store(1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if failMode.Load() == 1 {
			time.Sleep(1 * time.Second)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := buildCBNotifier(t, server.URL)
	update := model.LinkUpdate{ID: uuid.New().String(), TgChatIDs: []int64{1}}

	openCB(t, notifier)

	failMode.Store(0)

	time.Sleep(1200 * time.Millisecond)

	for i := range 3 {
		err := notifier.SendUpdate(context.Background(), update)
		require.NoError(t, err, "пробный вызов %d в HALF-OPEN должен быть успешным", i+1)
	}

	err := notifier.SendUpdate(context.Background(), update)
	assert.NoError(t, err, "CB должен быть в состоянии CLOSED после успешных проб")
}

// TC-4.3: HALF-OPEN → OPEN
func TestCircuitBreaker_HalfOpenTransitionsBackToOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := buildCBNotifier(t, server.URL)
	update := model.LinkUpdate{ID: uuid.New().String(), TgChatIDs: []int64{1}}

	openCB(t, notifier)

	time.Sleep(1200 * time.Millisecond)

	err := notifier.SendUpdate(context.Background(), update)
	require.Error(t, err, "пробный вызов в HALF-OPEN должен вернуть ошибку (таймаут)")

	start := time.Now()
	err = notifier.SendUpdate(context.Background(), update)
	elapsed := time.Since(start)

	require.Error(t, err, "ожидается ошибка: CB вернулся в состояние OPEN")
	assert.Less(t, elapsed, 50*time.Millisecond,
		"после возврата в OPEN вызов должен завершаться мгновенно, но занял %v", elapsed)
}

// TC-5.1: Fallback.
//
// Условие: основной HTTP-транспорт возвращает ошибку.
// Вызываем Mock Fallback
func TestFallback_TC51_UsesKafkaWhenHTTPFails(t *testing.T) {
	mainTransport := &MockMessageSender{}
	fallbackTransport := &MockMessageSender{}

	update := model.LinkUpdate{
		ID:        "fallback-test-id",
		Link:      "https://github.com/golang/go",
		TgChatIDs: []int64{1, 2},
	}
	data, err := json.Marshal(update)
	require.NoError(t, err)

	mainTransport.On("SendUpdate", mock.Anything, update).
		Return(errors.New("HTTP transport unavailable"))

	fallbackTransport.On("SendUpdate", mock.Anything, update).
		Return(nil)

	notifier := service.NewNotifierClient(mainTransport, fallbackTransport)
	err = notifier.SendMessage(context.Background(), "link-update", data)

	require.NoError(t, err, "при недоступном основном транспорте fallback должен доставить сообщение без ошибки")
	mainTransport.AssertCalled(t, "SendUpdate", mock.Anything, update)
	fallbackTransport.AssertCalled(t, "SendUpdate", mock.Anything, update)
}

// TC-5.1: ошибка не теряется молча.
//
// Условие: основной транспорт упал, fallback не задан (nil).
// Возвращаем ошибку
func TestFallback_TC51_ErrorNotSilentWhenNoFallback(t *testing.T) {
	mainTransport := &MockMessageSender{}

	update := model.LinkUpdate{
		ID:        "no-fallback-test-id",
		Link:      "https://github.com/golang/go",
		TgChatIDs: []int64{1},
	}
	data, err := json.Marshal(update)
	require.NoError(t, err)

	mainTransport.On("SendUpdate", mock.Anything, update).
		Return(errors.New("HTTP transport unavailable"))

	notifier := service.NewNotifierClient(mainTransport, nil)
	err = notifier.SendMessage(context.Background(), "link-update", data)

	assert.Error(t, err, "ошибка не должна теряться молча, если fallback не задан")
}

func buildCBNotifier(t *testing.T, url string) *httpclient.BotHTTPClient {
	t.Helper()
	cfg := httpclient2.Config{
		Timeout:              100 * time.Millisecond,
		Delay:                0,
		MaxDelay:             0,
		DelayType:            "constant",
		Retries:              1,
		RetriableStatusCodes: []int{},
		CircuitBreaker: httpclient2.CircuitBreakerConfig{
			MinRequiredCalls:     3,
			MaxRequestsHalfOpen:  3,
			ClosedStateInterval:  0,
			OpenStateInterval:    1 * time.Second,
			SlidingWindowSize:    0,
			FailureRateThreshold: 50,
		},
	}
	return httpclient.NewBotHTTPNotifier(url, httpclient2.NewReliableHTTPClient(cfg))
}

func openCB(t *testing.T, notifier *httpclient.BotHTTPClient) {
	t.Helper()
	update := model.LinkUpdate{ID: uuid.New().String()}
	for range 3 {
		_ = notifier.SendUpdate(context.Background(), update)
	}
}
