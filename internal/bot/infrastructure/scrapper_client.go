package infrastructure

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/dto"

type ScrapperClient interface {
	RegisterChat(chatID int64) error
	DeleteChat(chatID int64) error
	RegisterLink(chatID int64, request dto.AddLinkRequest) error
	DeleteLink(chatID int64, link string) error
	GetLinks(chatID int64) ([]dto.LinkResponse, error)
}
