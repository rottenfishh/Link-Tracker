package out

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

// TODO: обобщить интерфейс? я не могу придумать, как сделать его общим
// чтобы оно работало и для in-memory и для реальной бд
type ChatRepository interface {
	SaveChat(chat *model.Chat) error
	GetChats() ([]model.Chat, error)
	DeleteChat(chatId int64) error
	GetLinksById(chatId int64) ([]model.Link, error)
	GetFilteredLinksByTag(chatId int64, tag string) ([]model.Link, error)
	AddLink(chatId int64, link model.Link) (*model.Link, error)
	UpdateLink(chatId int64, link model.Link) (*model.Link, error)
	DeleteLink(chatId int64, linkName string) (*model.Link, error)
}

type LinkRepository interface {
	AddLink(link model.Link) (model.Link, error)
	DeleteLink(linkName string) error
}

type SubscriberRepository interface {
	Subscribe(chatId int64, link string) error
	Unsubscribe(chatId int64, link string) error
	GetSubscribers(link string) ([]model.Chat, error)
}
