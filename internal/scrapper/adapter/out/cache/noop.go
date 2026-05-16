package cache

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type NoOpCache struct{}

func (n NoOpCache) Set(_ context.Context, _ int64, _ []model.Link) error {
	return nil
}

func (n NoOpCache) Get(_ context.Context, _ int64) ([]model.Link, error) {
	return nil, model.ErrNotFound
}

func (n NoOpCache) Invalidate(_ context.Context, _ int64) error {
	return nil
}
