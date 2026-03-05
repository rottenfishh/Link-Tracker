package application

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/dto"

// TODO: не понимаю как все устроить чтобы было чисто архитектурно
type ScrapperClient interface {
	RegisterChat(chatID int64) error
	DeleteChat(chatID int64) error
	RegisterLink(chatID int64, request dto.AddLinkRequest) error
	DeleteLink(chatID int64, link dto.DeleteLinkRequest) error
	GetLinks(chatID int64) ([]dto.LinkResponse, error)
}
