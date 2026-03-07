package out

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

// TODO: what to do with errors in grpc?
type ScrapperClient interface {
	RegisterChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
	RegisterLink(ctx context.Context, request dto.AddLinkRequest) error
	DeleteLink(ctx context.Context, request dto.DeleteLinkRequest) error
	GetLinks(ctx context.Context, chatID int64) ([]dto.LinkResponse, error)
}
