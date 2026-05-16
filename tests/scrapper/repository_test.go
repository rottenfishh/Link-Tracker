package scrapper_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type TestEnv struct {
	DB              *sql.DB
	DSN             string
	ChatService     *service.ChatService
	TrackingService *service.LinkTrackingService
	Repos           *out.Repositories
	Trackers        []service.LinkUpdater
}

func TestChatService_IntegrationSuites(t *testing.T) {
	for _, accessType := range []string{"sql", "query"} {
		t.Run(accessType, func(t *testing.T) {
			env := SetUpTestEnv(t, accessType, http.DefaultClient)
			runChatServiceSuite(t, env)
		})
	}
}

func runChatServiceSuite(t *testing.T, env *TestEnv) {
	ctx := context.Background()

	t.Run("Добавление ссылки", func(t *testing.T) {
		chatID := int64(101)

		_, err := env.ChatService.RegisterChat(ctx, chatID)
		require.NoError(t, err)

		req := dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: []string{"go", "lang"},
		}

		added, err := env.ChatService.AddLink(ctx, chatID, req)
		require.NoError(t, err)
		require.NotNil(t, added)

		links, err := env.ChatService.GetLinksByChatID(ctx, chatID)
		require.NoError(t, err)
		require.Len(t, links, 1)
		require.Equal(t, req.Link, links[0].Link)

		require.True(t, subscriptionExists(t, env.DB, chatID, links[0].ID))
	})

	t.Run("Удаление ссылки", func(t *testing.T) {
		chatID := int64(102)

		_, err := env.ChatService.RegisterChat(ctx, chatID)
		require.NoError(t, err)

		req := dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: nil,
		}

		added, err := env.ChatService.AddLink(ctx, chatID, req)
		require.NoError(t, err)

		_, err = env.ChatService.Unsubscribe(ctx, chatID, dto.DeleteLinkRequest{
			Link: req.Link,
		})
		require.NoError(t, err)

		links, err := env.ChatService.GetLinksByChatID(ctx, chatID)
		require.NoError(t, err)
		require.Empty(t, links)

		require.False(t, subscriptionExists(t, env.DB, chatID, added.ID))
	})

	t.Run("Добавление дублирующей ссылки", func(t *testing.T) {
		chatID := int64(103)

		_, err := env.ChatService.RegisterChat(ctx, chatID)
		require.NoError(t, err)

		req := dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: []string{"go"},
		}

		_, err = env.ChatService.AddLink(ctx, chatID, req)
		require.NoError(t, err)

		_, err = env.ChatService.AddLink(ctx, chatID, req)
		require.Error(t, err)
	})

	t.Run("Получение ссылок по тегу", func(t *testing.T) {
		chatID := int64(104)

		_, err := env.ChatService.RegisterChat(ctx, chatID)
		require.NoError(t, err)

		_, err = env.ChatService.AddLink(ctx, chatID, dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: []string{"go"},
		})
		require.NoError(t, err)

		_, err = env.ChatService.AddLink(ctx, chatID, dto.AddLinkRequest{
			Link: "https://stackoverflow.com/questions/11227809/why-is-processing-a-sorted-array-faster-than-an-unsorted-array",
			Tags: []string{"so"},
		})
		require.NoError(t, err)

		links, err := env.ChatService.GetLinksByChatIDAndTag(ctx, chatID, "go")
		require.NoError(t, err)
		require.Len(t, links, 1)
		require.Equal(t, "https://github.com/golang/go", links[0].Link)
	})
}

func TestBuildRepos_UsesCorrectImplementation(t *testing.T) {
	for _, tc := range []struct {
		accessType string
		expected   string
	}{
		{accessType: "sql", expected: "sql."},
		{accessType: "query", expected: "querybuilder"},
	} {
		t.Run(tc.accessType, func(t *testing.T) {
			env := SetUpTestEnv(t, tc.accessType, http.DefaultClient)

			linkRepoType := reflect.TypeOf(env.Repos.LinkRepo).String()
			chatRepoType := reflect.TypeOf(env.Repos.ChatRepo).String()

			require.True(
				t,
				strings.Contains(linkRepoType, tc.expected) || strings.Contains(chatRepoType, tc.expected),
				"unexpected repo types: link=%s chat=%s. expected %s %s",
				linkRepoType,
				chatRepoType,
				tc.expected,
			)
		})
	}
}

func TestRunMigrations_OnCleanDB(t *testing.T) {
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
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/link-tracker?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.Eventually(t, func() bool {
		return db.PingContext(ctx) == nil
	}, 10*time.Second, 200*time.Millisecond)

	err = app.RunMigrations(dsn)
	require.NoError(t, err)
	expectedTables := []string{
		"chats",
		"links",
		"chat_link",
		"tags",
		"link_tag",
	}

	for _, tableName := range expectedTables {
		var exists bool
		queryErr := db.QueryRowContext(
			ctx,
			`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
			`,
			tableName,
		).Scan(&exists)
		require.NoError(t, queryErr)
		require.True(t, exists, "table %s does not exist after migrations", tableName)
	}
}

func subscriptionExists(t *testing.T, db *sql.DB, chatID, linkID int64) bool {
	t.Helper()

	var exists bool
	err := db.QueryRow(
		`
		SELECT EXISTS (
			SELECT 1
			FROM chat_link
			WHERE chat_ID = $1 AND link_ID = $2
		)
		`,
		chatID,
		linkID,
	).Scan(&exists)
	require.NoError(t, err)

	return exists
}

func SetUpTestEnv(t *testing.T, accessType string, client *http.Client) *TestEnv {
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

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/link-tracker?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	require.Eventually(t, func() bool {
		return db.PingContext(ctx) == nil
	}, 10*time.Second, 200*time.Millisecond)

	err = app.RunMigrations(dsn)
	require.NoError(t, err)

	repos, err := app.BuildRepos(accessType, db)
	require.NoError(t, err)

	trackers := []service.LinkUpdater{
		apiclients.NewGithubClient("mock-token", client),
		apiclients.NewStackOverflowClient("mock-token", client),
	}
	linkResolver := service.NewLinkResolver(trackers)
	txManager := repository.NewTransactionManager(db)

	cacheClient := cache.NoOpCache{}
	chatService := service.NewChatService(*repos, linkResolver, txManager, cacheClient)
	trackingService := service.NewLinkTrackingService(repos, txManager)

	return &TestEnv{
		DB:              db,
		DSN:             dsn,
		ChatService:     chatService,
		Repos:           repos,
		Trackers:        trackers,
		TrackingService: trackingService,
	}
}
