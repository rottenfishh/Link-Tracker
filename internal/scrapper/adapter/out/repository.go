package out

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

// TODO: обобщить интерфейс? я не могу придумать, как сделать его общим
// чтобы оно работало и для in-memory и для реальной бд
type ChatRepository interface {
	SaveChat(ctx context.Context, chat *model.Chat) error
	GetChats(ctx context.Context) ([]model.Chat, error)
	DeleteChat(ctx context.Context, chatId int64) (*model.Chat, error)
}

type LinkRepository interface {
	AddLink(ctx context.Context, link model.Link) (*model.Link, error)
	DeleteLink(ctx context.Context, linkID int64) (*model.Link, error)
	DeleteLinkByName(ctx context.Context, linkName string) (*model.Link, error)
	UpdateLink(ctx context.Context, linkID int64, link *model.Link) (*model.Link, error)
	GetLinks(ctx context.Context) ([]model.Link, error)
	GetLinkByName(ctx context.Context, linkName string) (*model.Link, error)
}

type ChatLinkRepository interface {
	GetLinksByChatID(ctx context.Context, chatId int64) ([]model.Link, error)
	Subscribe(ctx context.Context, chatId int64, linkId int64) error
	Unsubscribe(ctx context.Context, chatId int64, linkId int64) error
	GetChatsByLinkID(ctx context.Context, linkID int64) ([]model.Chat, error)
	GetChatIdsByLinkID(ctx context.Context, linkId int64) ([]int64, error)
}

type TagRepository interface {
	SaveTag(ctx context.Context, tag *model.Tag) (*model.Tag, error)
	GetTags(ctx context.Context) ([]model.Tag, error)
	DeleteTag(ctx context.Context, tagId int64) (*model.Tag, error)
	UpdateTag(ctx context.Context, tagId int64, tag *model.Tag) (*model.Tag, error)
	DeleteTagByName(ctx context.Context, tagName string) (*model.Tag, error)
}

type LinkTagRepository interface {
	GetLinksByTagID(ctx context.Context, tagId int64) ([]model.Link, error)
	Save(ctx context.Context, linkId int64, tagId int64) error
	Delete(ctx context.Context, linkId int64, tagId int64) error
}
