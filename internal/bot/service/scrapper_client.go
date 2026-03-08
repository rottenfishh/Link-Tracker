package service

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

// TODO: не понимаю как все устроить чтобы было чисто архитектурно
type ScrapperClient interface {
	RegisterChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
	RegisterLink(ctx context.Context, chatID int64, request dto.AddLinkRequest) error
	DeleteLink(ctx context.Context, chatID int64, link dto.DeleteLinkRequest) error
	GetLinks(ctx context.Context, chatID int64) (*dto.ListLinksResponse, error)
}
