//nolint:mnd // HTTP status codes are clearer inline in this small adapter
package out

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type ScrapperHTTPClient struct {
	url    string
	client httpclient.HTTPClient
}

func NewScrapperClient(url string, client httpclient.HTTPClient) *ScrapperHTTPClient {
	return &ScrapperHTTPClient{url: url, client: client}
}

func (c *ScrapperHTTPClient) RegisterChat(ctx context.Context, chatID int64) error {
	urlPath := c.url + "/tg-chat/" + strconv.FormatInt(chatID, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlPath, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("creating register chat request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending register chat request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing register chat response body", "error", closeErr)
		}
	}()
	// TODO: log docs error reponse
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *ScrapperHTTPClient) DeleteChat(ctx context.Context, chatID int64) error {
	urlPath := c.url + "/tg-chat/" + strconv.FormatInt(chatID, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlPath, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("creating delete chat request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending delete chat request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing delete chat response body", "error", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *ScrapperHTTPClient) RegisterLink(ctx context.Context, chatID int64, link string, tags []string) error {
	urlPath := c.url + "/links/" + strconv.FormatInt(chatID, 10)

	request := dto.AddLinkRequest{
		Link: link,
		Tags: tags,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshalling register link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlPath, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating register link request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending register link request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing register link response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var errResp dto.APIErrorResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&errResp)
		if decodeErr != nil {
			slog.Error("Unexpected error response", "error", decodeErr)
		}
		slog.Error("Error with code", "code", resp.StatusCode, "Error Response", errResp)

		switch resp.StatusCode {
		case 400:
			return model.ErrInvalidRequest
		case 404:
			return model.ErrNotFound
		case 409:
			return model.ErrLinkAlreadyTracked
		default:
			return model.ErrInternalServer
		}
	}
	return nil
}

func (c *ScrapperHTTPClient) DeleteLink(ctx context.Context, chatID int64, link string) error {
	urlPath := c.url + "/links/" + strconv.FormatInt(chatID, 10)

	deleteReq := dto.DeleteLinkRequest{Link: link}
	body, err := json.Marshal(deleteReq)
	if err != nil {
		return fmt.Errorf("marshalling delete link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlPath, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating delete link request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending delete link request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing delete link response body", "error", closeErr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 400:
			return model.ErrInvalidRequest
		case 404:
			return model.ErrNotFound
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *ScrapperHTTPClient) GetLinks(ctx context.Context, chatID int64, tag string) (*dto.ListLinksResponse, error) {
	urlPath := c.url + "/links/" + strconv.FormatInt(chatID, 10)
	if tag != "" {
		urlPath += "?tag=" + tag
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating get links request: %w", err)
	}
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(chatID, 10))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending get links request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing get links response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 400:
			return nil, model.ErrInvalidRequest
		case 404:
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var links dto.ListLinksResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&links); decodeErr != nil {
		return nil, fmt.Errorf("decoding get links response: %w", decodeErr)
	}
	return &links, nil
}
