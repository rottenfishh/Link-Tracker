package repository

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type InMemoryRepo struct {
	Chats      map[int64]*model.Chat
	nextLinkID int64
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{make(map[int64]*model.Chat), 0}
}
func (r *InMemoryRepo) SaveChat(chat *model.Chat) error {
	r.Chats[chat.ChatID] = chat
	return nil
}

func (r *InMemoryRepo) DeleteChat(chatID int64) error {
	delete(r.Chats, chatID)
	return nil
}

func (r *InMemoryRepo) GetChats() ([]model.Chat, error) {
	chats := make([]model.Chat, 0, len(r.Chats))
	for _, chat := range r.Chats {
		chats = append(chats, *chat)
	}
	return chats, nil
}
