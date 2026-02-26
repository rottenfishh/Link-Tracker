package application

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

func ParseLink(link string) string {
	parts := strings.Split(link, "/")
	parts = parts[:2]
	newLink := "https://api.github.com/repos/" + strings.Join(parts, "/")
	slog.Info(newLink)
	return newLink
}

func (c *GithubClient) GetUpdates(link string) (*http.Response, error) {
	//repo.getLink(link)
	newLink := ParseLink(link)
	req, err := http.NewRequest("GET", newLink, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Get("Last-Modified")
	// if link.LastModified < req { update it and users}
	req.Header.Add("Authorization", "token "+c.token)

	result, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Println(result)
	return result, nil
}
