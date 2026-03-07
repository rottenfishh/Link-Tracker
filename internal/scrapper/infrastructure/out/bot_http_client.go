package out

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/domain"
)

type BotHttpClient struct {
	client *http.Client
	url    string
}

func NewBotHttpNotifier(url string) *BotHttpClient {
	return &BotHttpClient{http.DefaultClient, url}
}

// TODO: return response
func (n *BotHttpClient) SendUpdate(ctx context.Context, update domain.LinkUpdate) error {
	slog.Info("Sending update ", "chat ", update.TgChatIds, " url ", update.Url)
	body, err := json.Marshal(update)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", n.url+"/updates", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	slog.Info(resp.Status)
	return nil
}
