package out

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/domain"
	dto2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

type ScrapperHttpClient struct {
	client *http.Client
	url    string
}

func NewScrapperClient(url string) *ScrapperHttpClient {
	return &ScrapperHttpClient{http.DefaultClient, url}
}

// Вызывается при /start
func (c *ScrapperHttpClient) RegisterChat(chatID int64) error {
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
	// TODO: log docs error reponse
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *ScrapperHttpClient) DeleteChat(chatID int64) error {
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

func (c *ScrapperHttpClient) RegisterLink(chatID int64, request dto2.AddLinkRequest) error {
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
		var errResp dto2.ApiErrorResponse
		err := json.NewDecoder(do.Body).Decode(&errResp)
		if err != nil {
			slog.Error("Unexpected error response", err)
		}
		slog.Error("Error with code", "code", do.StatusCode, "Error Response", errResp)

		switch do.StatusCode {
		case 400:
			return domain.ErrInvalidRequest
		case 404:
			return domain.ErrNotFound
		case 409:
			return domain.ErrLinkAlreadyTracked
		}
	}
	return nil
}

func (c *ScrapperHttpClient) DeleteLink(chatID int64, link dto2.DeleteLinkRequest) error {
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
		switch do.StatusCode {
		case 400:
			return domain.ErrInvalidRequest
		case 404:
			return domain.ErrNotFound
		}
		return fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}
	return nil
}

func (c *ScrapperHttpClient) GetLinks(chatID int64) ([]dto2.LinkResponse, error) {
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
		switch do.StatusCode {
		case 400:
			return nil, domain.ErrInvalidRequest
		case 404:
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}

	var links []dto2.LinkResponse
	if err := json.NewDecoder(do.Body).Decode(&links); err != nil {
		return nil, err
	}
	return links, nil
}
