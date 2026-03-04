package infrastructure

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type BotHttpNotifier struct {
	client *http.Client
	url    string
}

func NewBotHttpNotifier(url string) *BotHttpNotifier {
	return &BotHttpNotifier{http.DefaultClient, url}
}

// TODO: return response
func (n *BotHttpNotifier) SendUpdate(update domain.LinkUpdate) error {
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
