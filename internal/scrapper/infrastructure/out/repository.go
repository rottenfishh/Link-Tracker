package out

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/domain"

// TODO: обобщить интерфейс? я не могу придумать, как сделать его общим
// чтобы оно работало и для in-memory и для реальной бд
type ChatRepository interface {
	SaveChat(chat domain.Chat) error
	DeleteChat(chatId int64) error
	GetLinksById(chatId int64) ([]domain.Link, error)
	AddLink(chatId int64, link domain.Link) (domain.Link, error)
	DeleteLink(chatId int64, linkName string) error
	GetChatsByLink(link string) ([]domain.Chat, error)
}

type LinkRepository interface {
	AddLink(link domain.Link) (domain.Link, error)
	DeleteLink(linkName string) error
}

type SubscriberRepository interface {
	Subscribe(chatId int64, link string) error
	Unsubscribe(chatId int64, link string) error
	GetSubscribers(link string) ([]domain.Chat, error)
}
