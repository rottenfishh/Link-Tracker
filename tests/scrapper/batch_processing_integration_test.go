// testlint:disable=testifylint/http-handler its ok to have require in tests
package scrapper_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

// Тест: Пакетная обработка
// Батч корректно формируется и обрабатывается
func TestScheduler_FormsBatchesWithConfiguredBatchSize(t *testing.T) {
	env := SetUpTestEnv(t, "sql", http.DefaultClient)
	ctx := context.Background()

	chatID := int64(401)
	_, err := env.ChatService.RegisterChat(ctx, chatID)
	require.NoError(t, err)

	oldLinks := seedOldLinks(ctx, t, env, chatID, 5)

	jobsChan := make(chan []model.Link, 10)
	scheduler, err := service.NewScheduler(env.TrackingService, jobsChan, 2, 10)
	require.NoError(t, err)
	defer func() { _ = scheduler.Shutdown() }()

	err = scheduler.StartScheduler(ctx, 1)
	require.NoError(t, err)

	firstBatch := waitForBatch(t, jobsChan)
	secondBatch := waitForBatch(t, jobsChan)
	thirdBatch := waitForBatch(t, jobsChan)

	require.Len(t, firstBatch, 2)
	require.Len(t, secondBatch, 2)
	require.Len(t, thirdBatch, 1)

	gotLinks := collectLinks(firstBatch, secondBatch, thirdBatch)
	require.ElementsMatch(t, oldLinks, gotLinks)
}

func seedOldLinks(ctx context.Context, t *testing.T, env *TestEnv, chatID int64, count int) []string {
	t.Helper()

	createdLinks := make([]string, 0, count)
	baseTime := time.Now().UTC().Add(-2 * time.Hour)

	for i := range count {
		rawLink := fmt.Sprintf("https://github.com/acme/repo-%d", i)
		added, err := env.ChatService.AddLink(ctx, chatID, dto.AddLinkRequest{Link: rawLink})
		require.NoError(t, err)

		added.LastUpdated = baseTime.Add(time.Duration(i) * time.Minute)
		_, err = env.ChatService.UpdateLink(ctx, added)
		require.NoError(t, err)

		createdLinks = append(createdLinks, rawLink)
	}

	return createdLinks
}

func TestLinkProcessor_ProcessLinks_PartialFailure_WithHTTPTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/test/ok-1/issues":
			if !strings.Contains(r.URL.RawQuery, "since=") {
				t.Errorf("missing since query: %s", r.URL.RawQuery)
			}
			if !strings.Contains(r.URL.RawQuery, "state=all") {
				t.Errorf("missing state query: %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprintln(w, `[
				{
					"id": 1,
					"title": "Issue ok-1",
					"body": "preview for ok-1",
					"created_at": "2026-04-09T12:30:00Z",
					"user": { "login": "alice" }
				}
			]`)
			if err != nil {
				slog.Error("error writing from mock api")
			}
		case "/repos/test/ok-2/issues":
			if !strings.Contains(r.URL.RawQuery, "since=") {
				t.Errorf("missing since query: %s", r.URL.RawQuery)
			}
			if !strings.Contains(r.URL.RawQuery, "state=all") {
				t.Errorf("missing state query: %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprintln(w, `[
				{
					"id": 2,
					"title": "Issue ok-2",
					"body": "preview for ok-2",
					"created_at": "2026-04-09T13:30:00Z",
					"user": { "login": "bob" }
				}
			]`)
			if err != nil {
				slog.Error("error writing from mock api")
			}
		case "/repos/test/fail/issues":
			http.Error(w, "upstream unavailable", http.StatusInternalServerError)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	env := SetUpTestEnv(t, "sql", server.Client())
	ctx := context.Background()

	chatID := int64(777)
	_, err := env.ChatService.RegisterChat(ctx, chatID)
	require.NoError(t, err)

	ok1, err := env.ChatService.AddLink(ctx, chatID, dto.AddLinkRequest{
		Link: "https://github.com/test/ok-1",
	})
	require.NoError(t, err)

	failLink, err := env.ChatService.AddLink(ctx, chatID, dto.AddLinkRequest{
		Link: "https://github.com/test/fail",
	})
	require.NoError(t, err)

	ok2, err := env.ChatService.AddLink(ctx, chatID, dto.AddLinkRequest{
		Link: "https://github.com/test/ok-2",
	})
	require.NoError(t, err)

	ghClient := apiclients.NewGithubClient("mock-token", server.Client())

	ok1.FormattedLink = server.URL + "/repos/test/ok-1"
	failLink.FormattedLink = server.URL + "/repos/test/fail"
	ok2.FormattedLink = server.URL + "/repos/test/ok-2"

	ok1.Domain = "github"
	failLink.Domain = "github"
	ok2.Domain = "github"

	_, err = env.ChatService.UpdateLink(ctx, ok1)
	require.NoError(t, err)
	_, err = env.ChatService.UpdateLink(ctx, failLink)
	require.NoError(t, err)
	_, err = env.ChatService.UpdateLink(ctx, ok2)
	require.NoError(t, err)

	processor := service.NewLinkProcessor(
		make(chan []model.Link, 1),
		env.TrackingService,
	)
	processor.RegisterUpdater("github", ghClient)

	processor.ProcessLinks(ctx, []model.Link{*ok1, *failLink, *ok2})

	items, err := env.TrackingService.GetPendingOutbox(ctx)
	require.NoError(t, err)
	require.Len(t, items, 3)

	var updateLinks []string
	var reports []model.Report

	for _, item := range items {
		switch item.MessageType {
		case "link-update":
			var upd model.LinkUpdate
			err = json.Unmarshal(item.Payload, &upd)
			require.NoError(t, err)
			updateLinks = append(updateLinks, upd.Link)

		case "report":
			var report model.Report
			err = json.Unmarshal(item.Payload, &report)
			require.NoError(t, err)
			reports = append(reports, report)

		default:
			t.Fatalf("unexpected message type: %s", item.MessageType)
		}
	}

	require.ElementsMatch(t,
		[]string{ok1.Link, ok2.Link},
		updateLinks,
	)

	require.Len(t, reports, 1)
	require.Equal(t, chatID, reports[0].ChatID)
	require.Len(t, reports[0].LinkUpdateErrors, 1)
	require.Equal(t, failLink.Link, reports[0].LinkUpdateErrors[0].Link)
}

func waitForBatch(t *testing.T, jobsChan <-chan []model.Link) []model.Link {
	t.Helper()

	select {
	case batch := <-jobsChan:
		return batch
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for scheduler batch")
		return nil
	}
}

func collectLinks(batches ...[]model.Link) []string {
	links := make([]string, 0)
	for _, batch := range batches {
		for _, link := range batch {
			links = append(links, link.Link)
		}
	}
	return links
}
