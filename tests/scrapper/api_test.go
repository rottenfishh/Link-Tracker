package scrapper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"
)

// сценарий | результат
// Mock GitHub API возвращает новый Issue
// | Scrapper формирует сообщение с названием, автором, превью
// Mock StackOverflow API возвращает новый ответ
// | Аналогично
// Mock внешнего API имитирует недоступность ресурса
// | Бот корректно обрабатывает эту ситуацию
// Текст превью длиннее 200 символов
// | Превью обрезается до 200 символов
// Пакетная обработка
// | Батч корректно формируется и обрабатывается
// Пакетная обработка: часть ссылок не удалось обработать
// | Ошибки изолированы, остальные ссылки обработаны

func TestGithubClient_GetUpdatesAndTruncate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/test/repo/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "since=") {
			t.Errorf("missing since query: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "state=all") {
			t.Errorf("missing state query: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		resp := []map[string]any{
			{
				"id":         1,
				"title":      "New issue",
				"body":       "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
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

	defer server.Close()

	ghClient := apiclients.NewGithubClient("mock-token", server.Client())

	updates, err := ghClient.GetUpdates(&model.Link{
		ID:            2,
		FormattedLink: server.URL + "/repos/test/repo",
		LastUpdated:   time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC),
	})

	// check the response
	if err != nil {
		t.Fatalf("Unexpected error getting updates: %v", err)
	}

	if len(updates) == 0 {
		t.Fatalf("No updates found")
	}

	update := updates[0]
	require.Equal(t, "Issue: New issue", update.Title)
	require.Equal(t, "alice", update.Author)
	descLen := len(update.Description)
	if descLen == 0 || descLen > 200 {
		t.Fatalf("Update description is wrong length: %d", descLen)
	}
}

func TestStackOverflowClient_GetUpdates_AnswerAndTruncate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/questions/123/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "fromdate=") {
			t.Errorf("missing fromdate query: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "site=stackoverflow") {
			t.Errorf("missing site query: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")

		resp := map[string]any{
			"items": []map[string]any{
				{
					"creation_date": 1712665800,
					"body": `Lorem ipsum dolor sit amet, consectetur adipiscing elit,
                    sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
                    Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi
                    ut aliquip ex ea commodo consequat.Duis aute irure dolor in reprehenderit
                    in voluptate velit esse cillum dolore eu fugiat nulla pariatur.`,
					"title": "Test question",
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
	defer server.Close()

	soClient := apiclients.NewStackOverflowClient("mock-token", server.Client())

	link := &model.Link{
		ID:            1,
		FormattedLink: server.URL + "/questions/123",
		Title:         "Test question",
		LastUpdated:   time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC),
	}

	updates, err := soClient.GetUpdates(link)
	require.NoError(t, err)
	require.NotEmpty(t, updates)

	update := updates[0]

	require.Contains(t, update.Title, "Вопрос:")
	require.Equal(t, "alice", update.Author)

	descLen := len(update.Description)
	require.True(t, descLen > 0 && descLen <= 230, "description length invalid: %d", descLen)
}
