package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type OutboxRelay struct {
	Sender  *NotifierClient
	Service *LinkTrackingService
}

func NewOutboxRelay(sender *NotifierClient, service *LinkTrackingService) *OutboxRelay {
	return &OutboxRelay{
		Sender:  sender,
		Service: service,
	}
}

func (r *OutboxRelay) Run(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second) //nolint:mnd //idc
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down OutboxRelay")
			return fmt.Errorf("%w", ctx.Err())
		case <-ticker.C:
			items, err := r.Service.GetPendingOutbox(ctx)
			if err != nil {
				slog.Error("failed to get pending outbox items", "error", err)
				continue
			}

			for _, item := range items {
				if err = ctx.Err(); err != nil {
					slog.Info("Shutting down OutboxRelay")
					return fmt.Errorf("%w", err)
				}

				err = r.Sender.SendMessage(ctx, item.MessageType, item.Payload)
				if err != nil {
					slog.Error("failed to handle message", "error", err)
					continue
				}

				_, err = r.Service.MarkOutboxDone(ctx, item.ID, time.Now().UTC())
				if err != nil {
					slog.Error("Failed to mark outbox done", "error", err)
				}
			}
		}
	}
}
