package application

import (
	"fmt"
	"net/http"
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
func (c *GithubClient) GetUpdateFromLink(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "token "+c.token)

	result, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Println(result)
	return result, nil
}
