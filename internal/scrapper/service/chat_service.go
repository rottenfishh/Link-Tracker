package service

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
)

type ChatService struct {
	repo         out.ChatRepository
	linkResolver *LinkResolver
}

func NewChatService(repo out.ChatRepository, linkResolver *LinkResolver) *ChatService {
	return &ChatService{repo: repo, linkResolver: linkResolver}
}

func (s *ChatService) RegisterChat(id int64) (*model.Chat, error) {
	chat := model.NewChat(id)
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

func (s *ChatService) GetChats() ([]model.Chat, error) {
	return s.repo.GetChats()
}

func (s *ChatService) GetLinksByChatId(id int64) ([]model.Link, error) {
	links, err := s.repo.GetLinksById(id)
	if err != nil {
		return nil, err
	}
	return links, err
}

func (s *ChatService) GetLinksByChatIdAndTag(id int64, tag string) ([]model.Link, error) {
	if tag != "" {
		return s.repo.GetFilteredLinksByTag(id, tag)
	} else {
		return s.repo.GetLinksById(id)
	}
}

func (s *ChatService) AddLink(chatId int64, req dto.AddLinkRequest) (*model.Link, error) {
	link := model.NewLink(req.Link, req.Tags)
	link, err := s.linkResolver.FormatLink(link)
	if err != nil {
		return nil, err
	}

	addedLink, err := s.repo.AddLink(chatId, *link)
	if err != nil {
		return nil, err
	}
	return addedLink, nil
}

func (s *ChatService) DeleteLink(chatId int64, req dto.DeleteLinkRequest) (*model.Link, error) {
	link, err := s.repo.DeleteLink(chatId, req.Link)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (s *ChatService) UpdateLink(chatId int64, link model.Link) (*model.Link, error) {
	addedLink, err := s.repo.UpdateLink(chatId, link)
	if err != nil {
		return nil, err
	}
	return addedLink, nil
}
