package infrastructure

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/domain"

type Notifier interface {
	SendUpdate(update domain.LinkUpdate) error
}
