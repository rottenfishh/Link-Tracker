package service

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkCache interface {
	Set(ctx context.Context, chatID int64, links []model.Link) error
	Get(ctx context.Context, chatID int64) ([]model.Link, error)
	Invalidate(ctx context.Context, chatID int64) error
}
