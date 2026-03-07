package out

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/domain"
)

type InMemoryRepo struct {
	Chats       map[int64]*domain.Chat
	Subscribers map[string][]int64
	nextLinkId  int64
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{make(map[int64]*domain.Chat), make(map[string][]int64), 0}
}
func (r *InMemoryRepo) SaveChat(chat *domain.Chat) error {
	r.Chats[chat.Id] = chat
	return nil
}

func (r *InMemoryRepo) DeleteChat(chatId int64) error {
	delete(r.Chats, chatId)
	return nil
}

func (r *InMemoryRepo) GetChats() ([]domain.Chat, error) {
	chats := make([]domain.Chat, 0, len(r.Chats))
	for _, chat := range r.Chats {
		chats = append(chats, *chat)
	}
	return chats, nil
}

func (r *InMemoryRepo) GetLinksById(chatId int64) ([]domain.Link, error) {
	item, ok := r.Chats[chatId]
	if !ok {
		return []domain.Link{}, domain.ErrNotFound
	}
	links := item.Links
	return links, nil
}

func (r *InMemoryRepo) AddLink(chatId int64, link domain.Link) (*domain.Link, error) {
	_, ok := r.Chats[chatId]
	if !ok {
		r.Chats[chatId] = domain.NewChat(chatId)
	}
	for _, elem := range r.Chats[chatId].Links {
		if elem.Link == link.Link {
			return nil, domain.ErrLinkAlreadyTracked
		}
	}
	r.nextLinkId++
	link.Id = r.nextLinkId

	r.Chats[chatId].Links = append(r.Chats[chatId].Links, link)

	r.Subscribers[link.Link] = append(r.Subscribers[link.Link], chatId)
	return &link, nil
}

func (r *InMemoryRepo) UpdateLink(chatId int64, link domain.Link) (*domain.Link, error) {
	_, ok := r.Chats[chatId]
	if !ok {
		return nil, domain.ErrNotFound
	}
	for i, savedLink := range r.Chats[chatId].Links {
		if savedLink.Id == link.Id {
			r.Chats[chatId].Links[i] = link
		}
	}
	return &link, nil
}

// TODO: create func
func (r *InMemoryRepo) DeleteLink(chatId int64, linkName string) (*domain.Link, error) {
	_, ok := r.Chats[chatId]
	if !ok {
		return nil, domain.ErrNotFound
	}
	var idx int
	var linkDeleted domain.Link
	for index, link := range r.Chats[chatId].Links {
		if link.Link == linkName {
			idx = index
			linkDeleted = link
			break
		}
	}
	r.Chats[chatId].Links = append(r.Chats[chatId].Links[:idx], r.Chats[chatId].Links[idx+1:]...)

	for index, id := range r.Subscribers[linkName] {
		if id == chatId {
			idx = index
		}
	}
	r.Subscribers[linkName] = append(r.Subscribers[linkName][:idx], r.Subscribers[linkName][idx+1:]...)
	return &linkDeleted, nil
}

func (r *InMemoryRepo) GetChatsByLink(link string) ([]int64, error) {
	chats := r.Subscribers[link]
	return chats, nil
}
