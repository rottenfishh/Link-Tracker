package out

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type ScrapperHttpClient struct {
	client *http.Client
	url    string
}

func NewScrapperClient(url string) *ScrapperHttpClient {
	return &ScrapperHttpClient{http.DefaultClient, url}
}

// TODO: use context?
func (c *ScrapperHttpClient) RegisterChat(ctx context.Context, chatID int64) error {
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

func (c *ScrapperHttpClient) DeleteChat(ctx context.Context, chatID int64) error {
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

func (c *ScrapperHttpClient) RegisterLink(ctx context.Context, chatID int64, request dto.AddLinkRequest) error {
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
		var errResp dto.ApiErrorResponse
		err := json.NewDecoder(do.Body).Decode(&errResp)
		if err != nil {
			slog.Error("Unexpected error response", "error", err)
		}
		slog.Error("Error with code", "code", do.StatusCode, "Error Response", errResp)

		switch do.StatusCode {
		case 400:
			return model.ErrInvalidRequest
		case 404:
			return model.ErrNotFound
		case 409:
			return model.ErrLinkAlreadyTracked
		}
	}
	return nil
}

// pass link in query now
func (c *ScrapperHttpClient) DeleteLink(ctx context.Context, chatID int64, link dto.DeleteLinkRequest) error {
	urlPath := c.url + "/links/" + strconv.FormatInt(chatID, 10) + "?link=" + link.Link

	req, err := http.NewRequest("DELETE", urlPath, nil)
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
			return model.ErrInvalidRequest
		case 404:
			return model.ErrNotFound
		}
		return fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}
	return nil
}

func (c *ScrapperHttpClient) GetLinks(ctx context.Context, chatID int64) (*dto.ListLinksResponse, error) {
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
			return nil, model.ErrInvalidRequest
		case 404:
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("unexpected status code: %d", do.StatusCode)
	}

	var links dto.ListLinksResponse
	if err := json.NewDecoder(do.Body).Decode(&links); err != nil {
		return nil, err
	}
	return &links, nil
}
