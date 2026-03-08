package in

import (
	model2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type TgClient interface {
	SendMessage(chatID int64, message *model2.Message) error
	GetUpdates() <-chan model2.ChatUpdate
}
