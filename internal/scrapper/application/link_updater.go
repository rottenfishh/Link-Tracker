package application

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type LinkUpdater interface {
	FormatLink(link string) string
	GetUpdates(link string) (*domain.Update, error)
}
