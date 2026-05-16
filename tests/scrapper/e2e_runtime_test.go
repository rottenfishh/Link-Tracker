package scrapper_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	bot "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	http_out "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/tests/testenv"
)

func TestGithubUpdateFlow_Bot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/test/ok-1/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "since=") {
			t.Errorf("missing since query: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "state=all") {
			t.Errorf("missing state query: %s", r.URL.RawQuery)
		}

		resp := []map[string]any{
			{
				"id":         1,
				"title":      "New issue",
				"body":       "Github preview text",
				"created_at": "2026-04-09T12:30:00Z",
				"user": map[string]any{
					"login": "alice",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encoding github response: %v", err)
		}
	}))
	defer apiServer.Close()

	env := testenv.SetUpTestEnv(t, "sql", apiServer.Client())
	testChatID := int64(123)

	_, err := env.ChatService.RegisterChat(ctx, testChatID)
	require.NoError(t, err)

	link, err := env.ChatService.AddLink(ctx, testChatID, dto.AddLinkRequest{
		Link: "https://github.com/test/ok-1",
	})
	require.NoError(t, err)

	link.FormattedLink = apiServer.URL + "/repos/test/ok-1"
	//TODO: CREATE NORMAL LINK UPDATE
	_, err = env.ChatService.UpdateLink(ctx, link)
	require.NoError(t, err)

	cfg := &bot.AppConfig{
		Port:             8088,
		GatewayPort:      8080,
		ScrapperURL:      "http://localhost:8082",
		Telegram:         bot.TelegramConfig{},
		NThreads:         4,
		NotificationType: "grpc",
	}

	botApp, mockTg := testenv.BuildTestBot(ctx, cfg, nil, t)
	mockTg.On("GetUpdatesChan").Return(mockTg.Updates).Once()

	var sentMsg *model.Message
	mockTg.
		On("SendMessage", testChatID, mock.Anything).
		Run(func(args mock.Arguments) {
			sentMsg = args.Get(1).(*model.Message)
		}).
		Return(nil).
		Once()

	updates := make(chan model.ChatUpdate)
	close(updates)

	mockTg.On("GetUpdates").Return((<-chan model.ChatUpdate)(updates)).Maybe()
	go botApp.RunService(ctx)
	testenv.WaitForTCPPort(t, "127.0.0.1:8080")

	notifierClient := service.NewNotifierClient(
		http_out.NewBotHTTPNotifier("http://127.0.0.1:8080", http.DefaultClient),
		nil)

	dispatcher := service.NewOutboxRelay(
		notifierClient,
		env.TrackingService,
	)
	processor := service.NewLinkProcessor(make(chan []model.Link, 1), env.TrackingService)
	for _, tracker := range env.Trackers {
		processor.RegisterUpdater(tracker.GetDomain(), tracker)
	}

	go func() {
		err = dispatcher.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	}()

	go processor.ProcessLinks(ctx, []model.Link{*link})

	require.Eventually(t, func() bool {
		return sentMsg != nil
	}, 30*time.Second, 100*time.Millisecond)

	require.Contains(t, sentMsg.Text, "Обновление по ссылке: "+link.Link)
	require.Contains(t, sentMsg.Text, "New issue")
	require.Contains(t, sentMsg.Text, "alice")
	mockTg.AssertCalled(t, "SendMessage", testChatID, mock.Anything)
}

func TestStackOverflowUpdateFlow_Bot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/questions/123/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "fromdate=") {
			t.Errorf("missing fromdate query: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "site=stackoverflow") {
			t.Errorf("missing site query: %s", r.URL.RawQuery)
		}

		resp := map[string]any{
			"items": []map[string]any{
				{
					"creation_date": 1712665800,
					"body":          "StackOverflow preview text",
					"title":         "Test question",
					"owner": map[string]any{
						"display_name": "alice",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encoding stackoverflow response: %v", err)
		}
	}))
	defer apiServer.Close()

	env := testenv.SetUpTestEnv(t, "sql", apiServer.Client())
	testChatID := int64(123)

	_, err := env.ChatService.RegisterChat(ctx, testChatID)
	require.NoError(t, err)

	link, err := env.ChatService.AddLink(ctx, testChatID, dto.AddLinkRequest{
		Link: "https://stackoverflow.com/questions/123/test-question",
	})
	require.NoError(t, err)

	link.FormattedLink = apiServer.URL + "/questions/123"
	_, err = env.TrackingService.UpdateLinkLastModifiedOutbox(ctx, link, &model.LinkUpdate{
		ID:        uuid.New().String(),
		Link:      "https://stackoverflow.com/questions/123",
		Update:    model.Update{},
		TgChatIDs: nil,
	})
	require.NoError(t, err)

	cfg := &bot.AppConfig{
		Port:             8088,
		GatewayPort:      8080,
		ScrapperURL:      "http://localhost:8082",
		Telegram:         bot.TelegramConfig{},
		NThreads:         4,
		NotificationType: "grpc",
	}

	botApp, mockTg := testenv.BuildTestBot(ctx, cfg, nil, t)
	updates := make(chan model.ChatUpdate)
	close(updates)

	mockTg.On("GetUpdates").Return((<-chan model.ChatUpdate)(updates)).Maybe()

	var sentMsg *model.Message
	mockTg.
		On("SendMessage", testChatID, mock.Anything).
		Run(func(args mock.Arguments) {
			sentMsg = args.Get(1).(*model.Message)
		}).
		Return(nil).
		Twice()

	go botApp.RunService(ctx)
	testenv.WaitForTCPPort(t, "127.0.0.1:8080")

	notifierClient := service.NewNotifierClient(
		http_out.NewBotHTTPNotifier("http://127.0.0.1:8080", http.DefaultClient),
		nil)

	dispatcher := service.NewOutboxRelay(
		notifierClient,
		env.TrackingService,
	)
	processor := service.NewLinkProcessor(make(chan []model.Link, 1), env.TrackingService)
	for _, tracker := range env.Trackers {
		processor.RegisterUpdater(tracker.GetDomain(), tracker)
	}

	go func() {
		err = dispatcher.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	}()
	go processor.ProcessLinks(ctx, []model.Link{*link})

	require.Eventually(t, func() bool {
		return sentMsg != nil
	}, 30*time.Second, 100*time.Millisecond)

	require.Contains(t, sentMsg.Text, "Обновление по ссылке: "+link.Link)
	require.Contains(t, sentMsg.Text, "StackOverflow preview text")
	require.Contains(t, sentMsg.Text, "alice")
	mockTg.AssertCalled(t, "SendMessage", testChatID, mock.Anything)
}
