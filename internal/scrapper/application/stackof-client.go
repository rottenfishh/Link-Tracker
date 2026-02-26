package application

import (
	"fmt"
	"net/http"
)

type StackOverflowClient struct {
	client http.Client
	token  string
}

func (c *StackOverflowClient) GetUpdates(link string) (*http.Response, error) {
	//repo.getLink(link)
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+c.token)

	result, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	//result.Body.Read()
	// if link.LastModified < req { update it and users}

	fmt.Println(result)
	return result, nil
}
