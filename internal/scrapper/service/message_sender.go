package service

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type MessageSender interface {
	SendUpdate(ctx context.Context, update model.LinkUpdate) error
	SendReport(ctx context.Context, report model.Report) error
}
