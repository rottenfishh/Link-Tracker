package testenv

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in"
	grpcbot "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/kafka"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/out/repository"
	bot "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type BotDBEnv struct {
	DB   *sql.DB
	DSN  string
	Repo service.ProcessedEventsRepository
}

func SetUpBotDB(t *testing.T) *BotDBEnv {
	t.Helper()

	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "bot-db",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/bot-db?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	require.Eventually(t, func() bool {
		return db.PingContext(ctx) == nil
	}, 10*time.Second, 200*time.Millisecond) //nolint:mnd //its ok

	err = bot.RunMigrations(dsn)
	require.NoError(t, err)

	processedEventsRepo := repository.NewProcessedEventsRepository(db)
	require.NoError(t, err)

	return &BotDBEnv{
		DB:   db,
		DSN:  dsn,
		Repo: processedEventsRepo,
	}
}

func BuildTestBot(_ context.Context, cfg *bot.AppConfig, dbEnv *BotDBEnv, _ *testing.T) (*bot.App, *in.MockTgAdapter) {
	publisher := service.NewUpdatePublisher(100, 100) //nolint:mnd //its ok
	scrapperSvc := out.NewScrapperClient(cfg.ScrapperURL, http.DefaultClient)
	dispatcher := bot.BuildDispatcher(scrapperSvc)
	var repo service.ProcessedEventsRepository
	if dbEnv != nil {
		repo = repository.NewProcessedEventsRepository(dbEnv.DB)
	}

	mockTg := in.MockTgAdapter{
		Mock:    mock.Mock{},
		Updates: make(<-chan model.ChatUpdate),
	}

	var consumer in.UpdatesConsumer

	switch cfg.NotificationType {
	case "grpc":
		consumer = grpcbot.NewBotServiceServer(publisher, strconv.Itoa(cfg.Port), strconv.Itoa(cfg.GatewayPort))
	default:
		consumer = kafka.BuildKafkaConsumer(cfg.KafkaConfig, publisher, repo)
	}

	workers := bot.BuildUpdateWorkerPool(*cfg, publisher, &mockTg)
	botSvc := bot.NewApp(dispatcher, &mockTg, consumer, cfg, workers, repo)
	return botSvc, &mockTg
}

func WaitForTCPPort(t *testing.T, address string) {
	t.Helper()

	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond) //nolint:mnd, noctx //its ok
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 5*time.Second, 100*time.Millisecond) //nolint:mnd //its ok
}
