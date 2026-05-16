package scrapper_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	grpcin "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

var chatSeq int64 = 1000

func nextChatID() int64 {
	return atomic.AddInt64(&chatSeq, 1)
}

func TestScrapperIntegration(t *testing.T) {
	for _, accessType := range []string{"sql", "query"} {
		t.Run(accessType, func(t *testing.T) {
			baseURL := setupTestServer(t, accessType)
			runIntegrationSuite(t, baseURL)
		})
	}
}

func runIntegrationSuite(t *testing.T, baseURL string) {
	t.Run("AddAndGetLink", func(t *testing.T) {
		chatID := nextChatID()

		resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		addReq := dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: []string{"go", "lang"},
		}

		resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = http.Get(fmt.Sprintf("%s/links/%d", baseURL, chatID))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var linksResp dto.ListLinksResponse
		err = decodeJSON(resp, &linksResp)
		require.NoError(t, err)
		require.Len(t, linksResp.Links, 1)
		require.Equal(t, "https://github.com/golang/go", linksResp.Links[0].Link)
	})

	t.Run("AddAndDeleteLink", func(t *testing.T) {
		chatID := nextChatID()

		resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		addReq := dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: nil,
		}
		resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		delReq := dto.DeleteLinkRequest{
			Link: "https://github.com/golang/go",
		}
		resp, err = deleteJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), delReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = http.Get(fmt.Sprintf("%s/links/%d", baseURL, chatID))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var linksResp dto.ListLinksResponse
		err = decodeJSON(resp, &linksResp)
		require.NoError(t, err)
		require.Empty(t, linksResp.Links)
	})

	t.Run("DeleteLinkFromNonExistingChat", func(t *testing.T) {
		chatID := nextChatID()

		delReq := dto.DeleteLinkRequest{
			Link: "https://github.com/golang/go",
		}

		resp, err := deleteJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), delReq)
		require.NoError(t, err)
		require.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AddLinkToNonExistingChat", func(t *testing.T) {
		chatID := nextChatID()

		addReq := dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: nil,
		}

		resp, err := postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
		require.NoError(t, err)
		require.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeletedChatBehaviour", func(t *testing.T) {
		chatID := nextChatID()

		resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = deleteJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		addReq := dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: nil,
		}

		resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
		require.NoError(t, err)
		require.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteNonExistingChat", func(t *testing.T) {
		chatID := nextChatID()

		resp, err := deleteJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("AddInvalidLink", func(t *testing.T) {
		chatID := nextChatID()

		resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		addReq := dto.AddLinkRequest{
			Link: "bubblegum.com",
			Tags: nil,
		}

		resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), addReq)
		require.NoError(t, err)
		require.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetLinksByTag", func(t *testing.T) {
		chatID := nextChatID()

		resp, err := postJSON(fmt.Sprintf("%s/tg-chat/%d", baseURL, chatID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), dto.AddLinkRequest{
			Link: "https://github.com/golang/go",
			Tags: []string{"go"},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = postJSON(fmt.Sprintf("%s/links/%d", baseURL, chatID), dto.AddLinkRequest{
			Link: "https://stackoverflow.com/questions/11227809/why-is-processing-a-sorted-array-faster-than-an-unsorted-array",
			Tags: []string{"so"},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = http.Get(fmt.Sprintf("%s/links/%d?tag=go", baseURL, chatID))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var linksResp dto.ListLinksResponse
		err = decodeJSON(resp, &linksResp)
		require.NoError(t, err)
		require.Len(t, linksResp.Links, 1)
		require.Equal(t, "https://github.com/golang/go", linksResp.Links[0].Link)
	})
}

func setupTestServer(t *testing.T, accessType string) string {
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

	err = app.RunMigrations(dsn)
	require.NoError(t, err)

	repos, err := app.BuildRepos(accessType, db)
	require.NoError(t, err)

	trackers := []service.LinkUpdater{
		apiclients.NewGithubClient("mock-token", http.DefaultClient),
		apiclients.NewStackOverflowClient("mock-token", http.DefaultClient),
	}

	linkResolver := service.NewLinkResolver(trackers)
	txManager := repository.NewTransactionManager(db)
	cacheClient := cache.NoOpCache{}
	chatService := service.NewChatService(*repos, linkResolver, txManager, cacheClient)

	grpcPort := freePort(t)
	httpPort := freePort(t)
	baseURL := "http://127.0.0.1:" + httpPort

	server := grpcin.NewScrapperServer(chatService, &NoOpLimiter{}, grpcPort, httpPort)

	go func() {
		_ = server.RunServer(ctx)
	}()

	waitForHTTP(t, baseURL)

	return baseURL
}

type NoOpLimiter struct {
}

func (n *NoOpLimiter) Limit(next http.Handler) http.Handler {
	return next
}

func freePort(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func waitForHTTP(t *testing.T, baseURL string) {
	t.Helper()

	client := &http.Client{Timeout: 300 * time.Millisecond}

	require.Eventually(t, func() bool {
		resp, err := client.Get(baseURL + "/swagger/")
		if err == nil && resp != nil {
			_ = resp.Body.Close()
			return true
		}
		return false
	}, 10*time.Second, 200*time.Millisecond)
}

func postJSON(url string, body any) (*http.Response, error) {
	var buf *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(data)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	return http.Post(url, "application/json", buf)
}

func deleteJSON(url string, body any) (*http.Response, error) {
	var buf *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(data)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(http.MethodDelete, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func decodeJSON(resp *http.Response, dst any) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, dst)
}
