package apiclients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type GithubClient struct {
	token  string
	client httpclient.HTTPClient
}

const (
	previewMaxLength = 200
)

func NewGithubClient(token string, client httpclient.HTTPClient) *GithubClient {
	return &GithubClient{token: token, client: client}
}

func (c *GithubClient) GetTitle(formattedLink string) (string, error) {
	parts := strings.Split(strings.Trim(formattedLink, "/"), "/")
	repo := parts[len(parts)-1]
	return repo, nil
}

func (c *GithubClient) GetDomain() string {
	return "github"
}

func (c *GithubClient) FormatLink(link string) (string, error) {
	if !strings.Contains(link, "github") {
		return "", errors.New("invalid link")
	}
	u, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("parsing github link: %w", err)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	minGithubPathSegments := 2
	if len(parts) < minGithubPathSegments {
		return "", errors.New("invalid github url")
	}

	owner := parts[0]
	repo := parts[1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	slog.Debug("formatted github api url", "url", apiURL)
	return apiURL, nil
}

// TODO: map http error codes to errors
func (c *GithubClient) GetUpdates(link *model.Link) ([]model.Update, error) {
	reqPath := link.FormattedLink + "/issues?since=" + timeToIso(link.LastUpdated) + "&state=all"
	slog.Info("fetching github updates", "reqPath", reqPath)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating github request: %w", err)
	}
	req.Header.Add("Authorization", "token "+c.token)

	result, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending github request: %w", err)
	}
	defer func() {
		if closeErr := result.Body.Close(); closeErr != nil {
			slog.Error("closing github response body", "error", closeErr)
		}
	}()

	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", result.StatusCode)
	}

	resp, err := c.ParseResponse(result)
	if err != nil {
		slog.Error("Error parsing response", "error", err)
		return nil, fmt.Errorf("parsing github response: %w", err)
	}

	if len(resp) == 0 {
		slog.Debug("no updates in github")
		return nil, nil
	}

	timeModified := time.Now().UTC()

	var updates []model.Update
	for _, upd := range resp {
		update := c.FormatUpdate(timeModified, upd)
		updates = append(updates, update)
	}

	return updates, nil
}

func (c *GithubClient) ParseResponse(resp *http.Response) ([]model.GithubUpdate, error) {
	var githubUpdate []model.GithubUpdate
	if err := json.NewDecoder(resp.Body).Decode(&githubUpdate); err != nil {
		return nil, fmt.Errorf("decoding github response: %w", err)
	}
	return githubUpdate, nil
}

func (c *GithubClient) FormatUpdate(timeModified time.Time, upd model.GithubUpdate) model.Update {
	if upd.PullRequest != nil {
		upd.Type = "PullRequest"
	} else {
		upd.Type = "Issue"
	}

	preview := upd.Description
	if len(preview) > previewMaxLength {
		preview = preview[:previewMaxLength]
	}

	var update model.Update
	update.Title = upd.Type + ": " + upd.Title
	update.Author = upd.User.Login
	update.TimeCreated = upd.CreatedAt
	update.Description = preview
	update.LastModified = timeModified
	return update
}

func timeToIso(time time.Time) string {
	return time.Format("2006-01-02T15:04:05Z")
}
