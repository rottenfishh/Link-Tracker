// testifylint:disable:http-handler
package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	botkafka "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/kafka"
	bot "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	scrapperkafka "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/kafka"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/tests/testenv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

var kafkaEnv *testenv.KafkaSchemaRegistryEnv

func TestMain(m *testing.M) {
	var cleanup func()
	var err error
	kafkaEnv, cleanup, err = testenv.SetupKafkaWithSchemaRegistry(context.Background())
	if err != nil {
		fmt.Printf("error setting up bot_kafka schema registry: %v\n", err)
		os.Exit(1)
	}

	m.Run()
	cleanup()
}

func TestGithubUpdateFlow_WithKafka(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Log("set up bot_kafka")
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
		require.NoError(t, json.NewEncoder(w).Encode(resp)) //nolint:testifylint // its ok
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

	_, err = env.ChatService.UpdateLink(ctx, link)
	require.NoError(t, err)

	updatesTopic := "updates-github"
	reportsTopic := "report-github"

	botKafkaCfg := botkafka.Config{
		Brokers:           kafkaEnv.Brokers,
		SchemaURL:         kafkaEnv.SchemaRegistryURL,
		Topics:            []string{updatesTopic, reportsTopic},
		DeadLetterQueue:   "dead-letter-queue",
		GroupID:           "bot-test-" + uuid.NewString(),
		Offset:            "earliest",
		MinBytes:          1,
		MaxBytes:          10e6,
		MaxWaitMsecs:      100,
		MaxRetries:        3,
		ReplicationFactor: 1,
	}

	botCfg := &bot.AppConfig{
		Port:             8088,
		GatewayPort:      8080,
		ScrapperURL:      "http://localhost:8082",
		Telegram:         bot.TelegramConfig{},
		NThreads:         4,
		NotificationType: "bot_kafka",
		KafkaConfig:      botKafkaCfg,
	}

	botDBEnv := testenv.SetUpBotDB(t)

	botApp, mockTg := testenv.BuildTestBot(ctx, botCfg, botDBEnv, t)

	t.Log("built bot with bot_kafka")
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
		Once()
	t.Log("before running bot wiht bot_kafka")

	scrapperKafkaCfg := scrapperkafka.Config{
		SchemaURL:         kafkaEnv.SchemaRegistryURL,
		Brokers:           kafkaEnv.Brokers,
		UpdatesTopic:      updatesTopic,
		ReportsTopic:      reportsTopic,
		BatchSize:         100,
		BatchBytes:        1e6,
		BatchTimeoutMs:    100,
		RequiredAcks:      "all",
		MaxRetries:        3,
		ReplicationFactor: 1,
	}

	notifier := scrapperkafka.BuildKafkaProducer(scrapperKafkaCfg)

	notifierClient := service.NewNotifierClient(notifier, nil)
	dispatcher := service.NewOutboxRelay(
		notifierClient,
		env.TrackingService,
	)

	t.Log("built scrapper with bot_kafka")
	processor := service.NewLinkProcessor(make(chan []model.Link, 1), env.TrackingService)
	for _, tracker := range env.Trackers {
		processor.RegisterUpdater(tracker.GetDomain(), tracker)
	}

	go botApp.RunService(ctx)

	t.Log("run bot with bot_kafka")

	go func() {
		err = dispatcher.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	}()

	t.Log("run outbox with bot_kafka")
	go processor.ProcessLinks(ctx, []model.Link{*link})

	t.Log("run processor with bot_kafka")
	require.Eventually(t, func() bool {
		return sentMsg != nil
	}, 45*time.Second, 100*time.Millisecond)

	require.Contains(t, sentMsg.Text, "Обновление по ссылке: "+link.Link)
	require.Contains(t, sentMsg.Text, "New issue")
	require.Contains(t, sentMsg.Text, "alice")

	mockTg.AssertCalled(t, "SendMessage", testChatID, mock.Anything)
	t.Log("quitting")
}

func TestStackOverflowUpdateFlow_WithKafka(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("📡 Mock API request: %s %s", r.Method, r.URL.Path)

		if strings.Contains(r.URL.Path, "/questions/") {
			resp := map[string]any{
				"items": []map[string]any{
					{
						"question_id":        123,
						"title":              "Test question",
						"body":               "StackOverflow preview text",
						"creation_date":      1712665800,
						"last_activity_date": time.Now().Unix(),
						"owner": map[string]any{
							"display_name": "alice",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Errorf("Error encoding response: %v", err)
			}
			return
		}

		if strings.Contains(r.URL.Path, "/answers") {
			resp := map[string]any{
				"items": []map[string]any{
					{
						"answer_id":     456,
						"body":          "This is an answer",
						"creation_date": time.Now().Unix(),
						"owner": map[string]any{
							"display_name": "bob",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Errorf("Error encoding response: %v", err)
			}
			return
		}

		t.Errorf("❌ Unexpected path: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
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
	link, err = env.ChatService.UpdateLink(ctx, link)
	if err != nil {
		t.Error(err)
	}

	updatesTopic := "updates-stackof"
	reportsTopic := "report-stackof"

	botKafkaCfg := botkafka.Config{
		Brokers:           kafkaEnv.Brokers,
		SchemaURL:         kafkaEnv.SchemaRegistryURL,
		Topics:            []string{updatesTopic, reportsTopic},
		DeadLetterQueue:   "dead-letter-queue",
		GroupID:           "bot-test-" + uuid.NewString(),
		Offset:            "earliest",
		MinBytes:          1,
		MaxBytes:          10e6,
		MaxWaitMsecs:      100,
		MaxRetries:        3,
		ReplicationFactor: 1,
	}

	botCfg := &bot.AppConfig{
		Port:             8088,
		GatewayPort:      8080,
		ScrapperURL:      "http://localhost:8082",
		Telegram:         bot.TelegramConfig{},
		NThreads:         4,
		NotificationType: "bot_kafka",
		KafkaConfig:      botKafkaCfg,
	}

	botDBEnv := testenv.SetUpBotDB(t)

	botApp, mockTg := testenv.BuildTestBot(ctx, botCfg, botDBEnv, t)

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

	scrapperKafkaCfg := scrapperkafka.Config{
		SchemaURL:         kafkaEnv.SchemaRegistryURL,
		Brokers:           kafkaEnv.Brokers,
		UpdatesTopic:      updatesTopic,
		ReportsTopic:      reportsTopic,
		BatchSize:         100,
		BatchBytes:        1e6,
		BatchTimeoutMs:    100,
		RequiredAcks:      "all",
		MaxRetries:        3,
		ReplicationFactor: 1,
	}

	notifier := scrapperkafka.BuildKafkaProducer(scrapperKafkaCfg)

	notifierClient := service.NewNotifierClient(notifier, nil)
	dispatcher := service.NewOutboxRelay(
		notifierClient,
		env.TrackingService,
	)

	processor := service.NewLinkProcessor(make(chan []model.Link, 1), env.TrackingService)
	for _, tracker := range env.Trackers {
		processor.RegisterUpdater(tracker.GetDomain(), tracker)
	}
	go botApp.RunService(ctx)

	go func() {
		err = dispatcher.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	}()

	go processor.ProcessLinks(ctx, []model.Link{*link})
	t.Log("run processor with bot_kafka")
	require.Eventually(t, func() bool {
		return sentMsg != nil
	}, 30*time.Second, 100*time.Millisecond)

	require.Contains(t, sentMsg.Text, "Обновление по ссылке: "+link.Link)
	require.Contains(t, sentMsg.Text, "StackOverflow preview text")
	require.Contains(t, sentMsg.Text, "alice")
	mockTg.AssertCalled(t, "SendMessage", testChatID, mock.Anything)
}
