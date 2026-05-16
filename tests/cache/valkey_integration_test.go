package valkey_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	valkeyCache "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

var sharedValkeyAddr string

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "valkey/valkey:7.2-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("* Ready to accept connections").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		panic("failed to start valkey container: " + err.Error())
	}

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}
	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		panic(err)
	}
	sharedValkeyAddr = fmt.Sprintf("%s:%s", host, port.Port())

	code := m.Run()
	_ = container.Terminate(ctx)
	os.Exit(code)
}

type testEnv struct {
	chatService *service.ChatService
	cache       *valkeyCache.ValkeyClient
	mock        sqlmock.Sqlmock
}

func TestCacheStoresDataInExpectedFormat(t *testing.T) {
	ctx := context.Background()
	const chatID = int64(3001)
	const linkURL = "https://github.com/golang/go"

	env := buildEnv(t, 30*time.Second)
	expectGetLinksByChatID(env.mock, chatID, linkURL)

	links, err := env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, linkURL, links[0].Link)

	cached, err := env.cache.Get(ctx, chatID)
	require.NoError(t, err, "данные должны быть в кэше после первого GetLinksByChatID")
	require.Len(t, cached, 1)
	assert.Equal(t, linkURL, cached[0].Link)

	require.NoError(t, env.mock.ExpectationsWereMet())
}

func TestCacheHitDoesNotCallRepository(t *testing.T) {
	ctx := context.Background()
	const chatID = int64(3002)
	const linkURL = "https://github.com/golang/go"

	env := buildEnv(t, 30*time.Second)

	expectGetLinksByChatID(env.mock, chatID, linkURL)

	_, err := env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	// cache hit -> не должно быть sql запроса
	// Если сервис попытается обратиться к БД, sqlmock вернёт ошибку
	_, err = env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	_, err = env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	require.NoError(t, env.mock.ExpectationsWereMet())
}

// Проверяем, что 1 и 3 вызов приведут к кэш миссу, а 2 не приведет - т.к. ттл еще не выйдет
func TestCacheMissAfterTTLExpiry(t *testing.T) {
	ctx := context.Background()
	const chatID = int64(3003)
	const linkURL = "https://github.com/golang/go"
	const ttl = 1 * time.Second

	env := buildEnv(t, ttl)

	expectGetLinksByChatID(env.mock, chatID, linkURL)

	_, err := env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	_, err = env.chatService.GetLinksByChatID(ctx, chatID) // тут не должно быть кэш мисса
	require.NoError(t, err)

	time.Sleep(ttl + 500*time.Millisecond)
	expectGetLinksByChatID(env.mock, chatID, linkURL)

	_, err = env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	require.NoError(t, env.mock.ExpectationsWereMet())
}

func TestAddLinkInvalidatesCacheThenRepoIsCalled(t *testing.T) {
	ctx := context.Background()
	const chatID = int64(3004)
	const linkURL = "https://github.com/golang/go"

	env := buildEnv(t, 30*time.Second)

	expectGetLinksByChatID(env.mock, chatID, linkURL)
	expectAddLink(env.mock, chatID, linkURL)
	expectGetLinksByChatID(env.mock, chatID, linkURL)

	_, err := env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	cached, err := env.cache.Get(ctx, chatID)
	require.NoError(t, err)
	require.NotEmpty(t, cached)

	_, err = env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	_, err = env.chatService.AddLink(ctx, chatID, dto.AddLinkRequest{
		Link: linkURL,
		Tags: nil,
	})
	require.NoError(t, err)

	_, cacheErr := env.cache.Get(ctx, chatID)
	require.ErrorIs(t, cacheErr, model.ErrNotFound, "кэш должен быть инвалидирован после AddLink")

	_, err = env.chatService.GetLinksByChatID(ctx, chatID)
	require.NoError(t, err)

	require.NoError(t, env.mock.ExpectationsWereMet())
}

func buildEnv(t *testing.T, ttl time.Duration) *testEnv {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repos, err := app.BuildRepos("sql", db)
	require.NoError(t, err)

	cache, err := valkeyCache.NewValkeyClient(valkeyCache.Config{
		Addrs:              []string{sharedValkeyAddr},
		Timeout:            5 * time.Second,
		Expire:             ttl,
		UseClientSideCache: false,
	})
	require.NoError(t, err)

	txManager := repository.NewTransactionManager(db)
	linkResolver := service.NewLinkResolver([]service.LinkUpdater{
		apiclients.NewGithubClient("mock-token", httpclient.NewReliableHTTPClient(httpclient.Config{})),
	})

	svc := service.NewChatService(*repos, linkResolver, txManager, cache)
	return &testEnv{chatService: svc, cache: cache, mock: mock}
}
