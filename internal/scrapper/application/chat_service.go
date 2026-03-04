package application

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure/out"
)

type ChatService struct {
	repo out.ChatRepository
}

func NewChatService(repo out.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) RegisterChat(id int64) (*domain.Chat, error) {
	chat := domain.NewChat(id)
	err := s.repo.SaveChat(chat)
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatService) DeleteChat(id int64) error {
	err := s.repo.DeleteChat(id)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) GetChats() ([]domain.Chat, error) {
	return s.repo.GetChats()
}

func (s *ChatService) GetLinksByChatId(id int64) ([]domain.Link, error) {
	links, err := s.repo.GetLinksById(id)
	if err != nil {
		return nil, err
	}
	return links, err
}

func (s *ChatService) AddLink(chatId int64, req dto.AddLinkRequest) (*domain.Link, error) {
	link := domain.NewLink(req.Link, req.Tags)

	addedLink, err := s.repo.AddLink(chatId, *link)
	if err != nil {
		return nil, err
	}
	return addedLink, nil
}

func (s *ChatService) DeleteLink(chatId int64, req dto.DeleteLinkRequest) error {
	err := s.repo.DeleteLink(chatId, req.Link)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) UpdateLink(chatId int64, link domain.Link) (*domain.Link, error) {
	addedLink, err := s.repo.AddLink(chatId, link)
	if err != nil {
		return nil, err
	}
	return addedLink, nil
}
