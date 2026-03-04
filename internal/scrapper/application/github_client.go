package application

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type GithubClient struct {
	token  string
	client *http.Client
}

func NewGithubClient(token string) *GithubClient {
	client := http.DefaultClient
	return &GithubClient{token: token, client: client}
}

// response = requests.post(
// "https://api.github.com/user/repos",
// json=data,
// headers={"Authorization": f"token {token}"}
// )
//https://github.com/golang/go
//https://api.github.com/repos/golang/go

func (c *GithubClient) FormatLink(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid github url")
	}

	owner := parts[0]
	repo := parts[1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	slog.Info(apiURL)
	return apiURL, nil
}

func (c *GithubClient) GetUpdates(link string) (*domain.Update, error) {
	newLink, err := c.FormatLink(link)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", newLink, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+c.token)

	result, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	timeModified, err := parseTime(result.Header.Get("last-modified"))
	if err != nil {
		return nil, err
	}
	update := domain.NewUpdate(timeModified, "Update from github link "+link)
	slog.Info("github update: ", update)
	return update, nil
}

func parseTime(tm string) (time.Time, error) {
	if tm == "" {
		return time.Time{}, fmt.Errorf("no last-modified date found")
	}

	const layout = "Mon, 01 Jan 2006 15:04:05 MST"
	timeModified, err := time.Parse(layout, tm)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time from last modified header github api %v", err)
	}

	return timeModified, nil
}
