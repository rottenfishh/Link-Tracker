package application

import (
	dto2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

// TODO: не понимаю как все устроить чтобы было чисто архитектурно
type ScrapperClient interface {
	RegisterChat(chatID int64) error
	DeleteChat(chatID int64) error
	RegisterLink(chatID int64, request dto2.AddLinkRequest) error
	DeleteLink(chatID int64, link dto2.DeleteLinkRequest) error
	GetLinks(chatID int64) ([]dto2.LinkResponse, error)
}
