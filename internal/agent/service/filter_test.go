package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type mockProducer struct {
	mock.Mock
}

func (m *mockProducer) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

type mockSummarizer struct {
	mock.Mock
}

func (m *mockSummarizer) Summarize(_ context.Context, description string, length int) (string, error) {
	args := m.Called(description, length)
	return args.String(0), args.Error(1)
}

func defaultCfg() service.FilterConfig {
	cfg := service.FilterConfig{
		StopWords:     []string{"spam", "ads", "promo"},
		BannedAuthors: []string{"bot", "spammer"},
		MinLength:     10,
	}
	cfg.Summarization.Threshold = 100
	return cfg
}

func makeUpdate(description, author string) model.LinkUpdate {
	return model.LinkUpdate{
		ID:   "test-id",
		Link: "https://example.com",
		Update: model.Update{
			Title:       "Test title",
			Author:      author,
			Description: description,
		},
		TgChatIDs: []int64{1},
	}
}

// TC-2.1: Фильтрация по стоп-словам
func TestFilter_StopWord_BlocksUpdate(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	filter := service.NewFilter(defaultCfg(), summarizer, producer)

	upd := makeUpdate("this message contains spam and more text", "alice")
	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertNotCalled(t, "SendUpdate", mock.Anything, mock.Anything)
}

func TestFilter_StopWord_CaseInsensitive_BlocksUpdate(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	filter := service.NewFilter(defaultCfg(), summarizer, producer)

	upd := makeUpdate("SPAM detected in this update text", "alice")
	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertNotCalled(t, "SendUpdate", mock.Anything, mock.Anything)
}

func TestFilter_NoStopWords_AllowsUpdate(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	cfg := defaultCfg()
	cfg.Summarization.Threshold = 1000

	filter := service.NewFilter(cfg, summarizer, producer)

	upd := makeUpdate("this is a clean and valid update", "alice")
	producer.On("SendUpdate", mock.Anything, mock.Anything).Return(nil)

	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertCalled(t, "SendUpdate", mock.Anything, mock.Anything)
}

// TC-2.2: Фильтрация по автору
func TestFilter_BannedAuthor_BlocksUpdate(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	filter := service.NewFilter(defaultCfg(), summarizer, producer)

	upd := makeUpdate("this is a valid update text here", "bot")
	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertNotCalled(t, "SendUpdate", mock.Anything, mock.Anything)
}

func TestFilter_NonBannedAuthor_AllowsUpdate(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	cfg := defaultCfg()
	cfg.Summarization.Threshold = 1000

	filter := service.NewFilter(cfg, summarizer, producer)

	upd := makeUpdate("this is a valid update text here", "alice")
	producer.On("SendUpdate", mock.Anything, mock.Anything).Return(nil)

	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertCalled(t, "SendUpdate", mock.Anything, mock.Anything)
}

// TC-2.3: Фильтрация по минимальной длине
func TestFilter_TooShortDescription_BlocksUpdate(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	filter := service.NewFilter(defaultCfg(), summarizer, producer)

	upd := makeUpdate("short", "alice")
	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertNotCalled(t, "SendUpdate", mock.Anything, mock.Anything)
}

func TestFilter_ExactlyMinLength_AllowsUpdate(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	cfg := defaultCfg()
	cfg.MinLength = 5
	cfg.Summarization.Threshold = 1000

	filter := service.NewFilter(cfg, summarizer, producer)

	upd := makeUpdate("hello world valid update", "alice")
	producer.On("SendUpdate", mock.Anything, mock.Anything).Return(nil)

	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertCalled(t, "SendUpdate", mock.Anything, mock.Anything)
}

// TC-2.4: Обновление проходит все фильтры
func TestFilter_ValidUpdate_PassesAllFilters(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	cfg := service.FilterConfig{
		StopWords:     []string{"spam", "ads"},
		BannedAuthors: []string{"bot"},
		MinLength:     10,
	}
	cfg.Summarization.Threshold = 1000

	filter := service.NewFilter(cfg, summarizer, producer)

	upd := makeUpdate("this is a clean valid update message", "alice")
	producer.On("SendUpdate", mock.Anything, upd).Return(nil)

	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertExpectations(t)
}

func TestFilter_ValidUpdate_ProducerError_ReturnsError(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	cfg := defaultCfg()
	cfg.Summarization.Threshold = 1000

	filter := service.NewFilter(cfg, summarizer, producer)

	upd := makeUpdate("this is a clean valid update message", "alice")
	producer.On("SendUpdate", mock.Anything, mock.Anything).Return(errors.New("kafka unavailable"))

	err := filter.HandleUpdate(context.Background(), upd)

	require.Error(t, err)
}

// TC-3.1: Суммаризация длинного текста

func TestFilter_LongText_Summarized(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &service.StupidSummarizer{}

	cfg := service.FilterConfig{
		StopWords:     []string{},
		BannedAuthors: []string{},
		MinLength:     1,
	}
	cfg.Summarization.Threshold = 20

	filter := service.NewFilter(cfg, summarizer, producer)

	longDesc := "this is a very long description that exceeds the summarization threshold"
	shortSummary := "this is a very long"

	producer.On("SendUpdate", mock.Anything, mock.MatchedBy(func(u model.LinkUpdate) bool {
		return u.Update.Description == shortSummary && u.Update.Description != longDesc
	})).Return(nil)

	upd := makeUpdate(longDesc, "alice")
	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)
	producer.AssertExpectations(t)
}

// TC-3.2: Короткий текст — без суммаризации
func TestFilter_ShortText_NotSummarized(t *testing.T) {
	producer := &mockProducer{}
	summarizer := &mockSummarizer{}

	cfg := service.FilterConfig{
		StopWords:     []string{},
		BannedAuthors: []string{},
		MinLength:     1,
	}
	cfg.Summarization.Threshold = 100

	filter := service.NewFilter(cfg, summarizer, producer)

	shortDesc := "short update text"
	producer.On("SendUpdate", mock.Anything, mock.MatchedBy(func(u model.LinkUpdate) bool {
		return u.Update.Description == shortDesc
	})).Return(nil)

	upd := makeUpdate(shortDesc, "alice")
	err := filter.HandleUpdate(context.Background(), upd)

	require.NoError(t, err)

	summarizer.AssertNotCalled(t, "Summarize", mock.Anything, mock.Anything)
	producer.AssertExpectations(t)
}

// Stupid Summarizer tests
func TestStupidSummarizer_TruncatesToLength(t *testing.T) {
	s := &service.StupidSummarizer{}
	input := "hello world this is a long string"
	result, err := s.Summarize(context.Background(), input, 11)
	require.NoError(t, err)
	require.Equal(t, "hello world", result)
	require.Len(t, result, 11)
}

func TestStupidSummarizer_ShortInput_ReturnsPrefix(t *testing.T) {
	s := &service.StupidSummarizer{}
	input := "hi there"
	result, err := s.Summarize(context.Background(), input, 2)
	require.Equal(t, "hi", result)
	require.NoError(t, err)
}
