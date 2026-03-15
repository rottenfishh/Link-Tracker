package out

import (
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type InMemoryRepo struct {
	Chats      map[int64]*model.Chat
	nextLinkId int64
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{make(map[int64]*model.Chat), 0}
}
func (r *InMemoryRepo) SaveChat(chat *model.Chat) error {
	r.Chats[chat.Id] = chat
	return nil
}

func (r *InMemoryRepo) DeleteChat(chatId int64) error {
	delete(r.Chats, chatId)
	return nil
}

func (r *InMemoryRepo) GetChats() ([]model.Chat, error) {
	chats := make([]model.Chat, 0, len(r.Chats))
	for _, chat := range r.Chats {
		chats = append(chats, *chat)
	}
	return chats, nil
}

func (r *InMemoryRepo) GetLinksById(chatId int64) ([]model.Link, error) {
	item, ok := r.Chats[chatId]
	if !ok {
		return []model.Link{}, model.ErrNotFound
	}
	slog.Debug("got links in repo", "links", item.Links, "id", chatId)
	links := item.Links
	return links, nil
}

func (r *InMemoryRepo) AddLink(chatId int64, link model.Link) (*model.Link, error) {
	_, ok := r.Chats[chatId]
	if !ok {
		return nil, model.ErrNotFound
	}
	for _, elem := range r.Chats[chatId].Links {
		if elem.Link == link.Link {
			return nil, model.ErrLinkAlreadyTracked
		}
	}
	r.nextLinkId++
	link.Id = r.nextLinkId

	r.Chats[chatId].Links = append(r.Chats[chatId].Links, link)
	return &link, nil
}

func (r *InMemoryRepo) UpdateLink(chatId int64, link model.Link) (*model.Link, error) {
	_, ok := r.Chats[chatId]
	if !ok {
		return nil, model.ErrNotFound
	}
	for i, savedLink := range r.Chats[chatId].Links {
		if savedLink.Id == link.Id {
			r.Chats[chatId].Links[i] = link
		}
	}
	return &link, nil
}

func (r *InMemoryRepo) DeleteLink(chatId int64, linkName string) (*model.Link, error) {
	_, ok := r.Chats[chatId]
	if !ok {
		return nil, model.ErrNotFound
	}
	var idx int
	var linkDeleted model.Link
	for index, link := range r.Chats[chatId].Links {
		if link.Link == linkName {
			idx = index
			linkDeleted = link
			break
		}
	}
	r.Chats[chatId].Links = append(r.Chats[chatId].Links[:idx], r.Chats[chatId].Links[idx+1:]...)

	return &linkDeleted, nil
}
