package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type BotHTTPClient struct {
	url    string
	client httpclient.HTTPClient
}

func NewBotHTTPNotifier(url string, client httpclient.HTTPClient) *BotHTTPClient {
	return &BotHTTPClient{url: url, client: client}
}

func (n *BotHTTPClient) SendReport(ctx context.Context, report model.Report) error {
	slog.Info("Sending report", "report", report)
	body, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshalling report: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.url+"/reports", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating report request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending report request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code %d", resp.StatusCode)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing report response body", "error", closeErr)
		}
	}()

	return nil
}

// TODO: return response
func (n *BotHTTPClient) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	slog.Info("Sending update ", "chat ", update.TgChatIDs, " url ", update.Link, "update", update.Update)
	body, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("marshalling update: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.url+"/updates", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending update request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("closing update response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code %d", resp.StatusCode)
	}

	return nil
}
