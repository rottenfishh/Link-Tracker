package in

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type TgClient interface {
	SendMessage(chatID int64, message *model.Message) error
	GetUpdates() <-chan model.ChatUpdate
}
