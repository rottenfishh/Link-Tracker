package service

import (
	"context"
)

type ProcessedEventsRepository interface {
	Register(ctx context.Context, id string) (bool, error)
	Delete(ctx context.Context, id string) error
}
