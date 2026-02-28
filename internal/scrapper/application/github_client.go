package application

import (
	"fmt"
	"log/slog"
	"net/http"
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

func (c *GithubClient) FormatLink(link string) string {
	parts := strings.Split(link, "/")
	parts = parts[:2]
	newLink := "https://api.github.com/repos/" + strings.Join(parts, "/")
	slog.Info(newLink)
	return newLink
}

func (c *GithubClient) GetUpdates(link string) (*domain.Update, error) {
	newLink := c.FormatLink(link)
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
	return update, nil
}

func parseTime(tm string) (time.Time, error) {
	if tm == "" {
		return time.Time{}, fmt.Errorf("no last-modified date found")
	}

	const layout = "01/01 Mon, 2 Jan 2006 15:04:05 MST"
	timeModified, err := time.Parse(layout, tm)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time from last modified header github api %v", err)
	}

	return timeModified, nil
}
