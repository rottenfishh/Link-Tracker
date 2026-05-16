package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type NotifierClient struct {
	MainTransport     MessageSender
	FallBackTransport MessageSender
}

func NewNotifierClient(main MessageSender, fallback MessageSender) *NotifierClient {
	return &NotifierClient{MainTransport: main, FallBackTransport: fallback}
}

// TODO: убрать дублирование как то
func (c *NotifierClient) SendMessage(ctx context.Context, msgType string, message []byte) error {
	switch msgType {
	case "link-update":
		var update model.LinkUpdate
		err := json.Unmarshal(message, &update)
		if err != nil {
			return fmt.Errorf("failed to unmarshal update: %w", err)
		}

		err = c.MainTransport.SendUpdate(ctx, update)
		if err == nil {
			return nil
		}
		slog.Error("sending report to main transport failed", "error", err)
		slog.Info("calling fallback transport")
		if c.FallBackTransport == nil {
			return errors.New("main transport failed and no fallback transport set")
		}

		err = c.FallBackTransport.SendUpdate(ctx, update)
		if err != nil {
			return fmt.Errorf("sending update to fallback transport failed: %w", err)
		}
		return nil
	case "report":
		var report model.Report
		err := json.Unmarshal(message, &report)
		if err != nil {
			return fmt.Errorf("failed to unmarshal report: %w", err)
		}

		err = c.MainTransport.SendReport(ctx, report)
		if err == nil {
			return nil
		}
		slog.Error("sending report to main transport failed", "error", err)
		slog.Info("calling fallback transport")
		if c.FallBackTransport == nil {
			return errors.New("main transport failed and no fallback transport set")
		}

		err = c.FallBackTransport.SendReport(ctx, report)
		if err != nil {
			return fmt.Errorf("sending report to fallback transport failed: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported message type: %s", msgType)
	}
}
