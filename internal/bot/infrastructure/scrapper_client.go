package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/dto"
)

type ScrapperClient struct {
	client *http.Client
	url    string
}

func NewScrapperClient(url string) *ScrapperClient {
	return &ScrapperClient{http.DefaultClient, url}
}

// Вызывается при /start
func (c *ScrapperClient) RegisterChat(chatID int64) error {
	urlPath := c.url + "/tg-chat/" + strconv.FormatInt(chatID, 10)
	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// TODO: log api error reponse
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *ScrapperClient) DeleteChat(chatID int64) error {
	urlPath := c.url + "/tg-chat/" + strconv.FormatInt(chatID, 10)
	req, err := http.NewRequest("DELETE", urlPath, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}

	do, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}
	return nil
}

func (c *ScrapperClient) TrackLink(chatID int64, request dto.AddLinkRequest) error {
	urlPath := c.url + "/links/" + strconv.FormatInt(chatID, 10)

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	do, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}
	return nil
}

func (c *ScrapperClient) UntrackLink(chatID int64, link string) error {
	urlPath := c.url + "/links/" + strconv.FormatInt(chatID, 10)

	body, err := json.Marshal(link)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", urlPath, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	do, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}
	return nil
}

func (c *ScrapperClient) GetLinks(chatID int64) (*dto.LinkResponse, error) {
	urlPath := c.url + "/links/" + strconv.FormatInt(chatID, 10)

	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	do, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer do.Body.Close()

	if do.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}

	var link dto.LinkResponse
	if err := json.NewDecoder(do.Body).Decode(&link); err != nil {
		return nil, err
	}
	return &link, nil
}
