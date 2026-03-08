package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type BotHttpClient struct {
	client *http.Client
	url    string
}

func NewBotHttpNotifier(url string) *BotHttpClient {
	return &BotHttpClient{http.DefaultClient, url}
}

// TODO: return response
func (n *BotHttpClient) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	slog.Info("Sending update ", "chat ", update.TgChatIds, " url ", update.Link)
	body, err := json.Marshal(update)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", n.url+"/updates", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "service/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	slog.Info(resp.Status)
	return nil
}
