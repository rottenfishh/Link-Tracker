package in

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
)

type UpdatesConsumer interface {
	GetUpdates() chan service.UpdateJob
	GetReports() chan service.ReportJob
	Run(ctx context.Context) error
}
