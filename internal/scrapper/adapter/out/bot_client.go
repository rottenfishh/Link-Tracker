package out

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Notifier interface {
	SendUpdate(ctx context.Context, update model.LinkUpdate) error
}
