package out

import (
	domain2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type InMemoryRepo struct {
	Chats       map[int64]*domain2.Chat
	Subscribers map[string][]int64
}

func (r *InMemoryRepo) SaveChat(chat *domain2.Chat) error {
	r.Chats[chat.Id] = chat
	return nil
}

func (r *InMemoryRepo) DeleteChat(chatId int64) error {
	delete(r.Chats, chatId)
	return nil
}

func (r *InMemoryRepo) GetLinksById(chatId int64) ([]domain2.Link, error) {
	links := r.Chats[chatId].Links
	return links, nil
}

func (r *InMemoryRepo) AddLink(chatId int64, link domain2.Link) (domain2.Link, error) {
	r.Chats[chatId].Links = append(r.Chats[chatId].Links, link)

	r.Subscribers[link.Link] = append(r.Subscribers[link.Link], chatId)
	return link, nil
}

// TODO: create func
func (r *InMemoryRepo) DeleteLink(chatId int64, linkName string) error {
	var idx int
	for index, link := range r.Chats[chatId].Links {
		if link.Link == linkName {
			idx = index
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
	return nil
}

func (r *InMemoryRepo) GetChatsByLink(link string) ([]int64, error) {
	chats := r.Subscribers[link]
	return chats, nil
}
