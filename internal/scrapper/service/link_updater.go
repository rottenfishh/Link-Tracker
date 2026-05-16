package service

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

// TODO: if linkupdater not present, send message to user
type LinkUpdater interface {
	FormatLink(link string) (string, error)
	GetDomain() string
	GetTitle(link string) (string, error)
	GetUpdates(link *model.Link) ([]model.Update, error)
}
