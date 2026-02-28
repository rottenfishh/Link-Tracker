package out

import (
	domain2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

// TODO: обобщить интерфейс? я не могу придумать, как сделать его общим
// чтобы оно работало и для in-memory и для реальной бд
type ChatRepository interface {
	SaveChat(chat domain2.Chat) error
	DeleteChat(chatId int64) error
	GetLinksById(chatId int64) ([]domain2.Link, error)
	AddLink(chatId int64, link domain2.Link) (domain2.Link, error)
	DeleteLink(chatId int64, linkName string) error
	GetChatsByLink(link string) ([]domain2.Chat, error)
}

type LinkRepository interface {
	AddLink(link domain2.Link) (domain2.Link, error)
	DeleteLink(linkName string) error
}

type SubscriberRepository interface {
	Subscribe(chatId int64, link string) error
	Unsubscribe(chatId int64, link string) error
	GetSubscribers(link string) ([]domain2.Chat, error)
}
