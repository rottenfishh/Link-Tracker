package scrapper_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"
	grpcin "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/middleware"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

func setupRateLimitedServer(t *testing.T, rlCfg middleware.RateLimitConfig) string {
	t.Helper()
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "link-tracker",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/link-tracker?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	err = app.RunMigrations(dsn)
	require.NoError(t, err)

	repos, err := app.BuildRepos("sql", db)
	require.NoError(t, err)

	trackers := []service.LinkUpdater{
		apiclients.NewGithubClient("mock-token", http.DefaultClient),
		apiclients.NewStackOverflowClient("mock-token", http.DefaultClient),
	}
	linkResolver := service.NewLinkResolver(trackers)
	txManager := repository.NewTransactionManager(db)
	cacheClient := cache.NoOpCache{}
	chatService := service.NewChatService(*repos, linkResolver, txManager, cacheClient)

	rl := middleware.NewRateLimiter(rlCfg)

	grpcPort := freePort(t)
	httpPort := freePort(t)
	baseURL := "http://127.0.0.1:" + httpPort

	server := grpcin.NewScrapperServer(chatService, rl, grpcPort, httpPort)
	go func() { _ = server.RunServer(ctx) }()

	waitForHTTP(t, baseURL)
	return baseURL
}

// TC-3.1 Превышение лимита
func TestRateLimit_TC31_ExceedingRequestsGet429(t *testing.T) {
	const burst = 3

	baseURL := setupRateLimitedServer(t, middleware.RateLimitConfig{
		Burst:             burst,
		Limit:             1,
		VisitorExpiration: 5 * time.Minute,
	})

	chatID := nextChatID()

	resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"запрос 1 (в пределах лимита): регистрация чата должна вернуть 200")

	for i := 2; i < burst; i++ {
		resp, err = http.Get(fmt.Sprintf("%s/links/%d", baseURL, chatID))
		require.NoError(t, err)
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"запрос %d (в пределах лимита) должен вернуть 200, а не 429", i)
	}

	const extra = 5
	for i := 1; i <= extra; i++ {
		resp, err = http.Get(fmt.Sprintf("%s/links/%d", baseURL, chatID))
		require.NoError(t, err)
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode,
			"запрос %d сверх лимита должен вернуть 429", burst+i)
	}
}
