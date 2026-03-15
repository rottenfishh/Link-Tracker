package service

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
)

type ChatService struct {
	ChatRepo     out.ChatRepository
	LinkRepo     out.LinkRepository
	ChatLinkRepo out.ChatLinkRepository
}

func NewChatService(chatRepo out.ChatRepository, linkRepo out.LinkRepository, chatLinkRepo out.ChatLinkRepository) *ChatService {
	return &ChatService{chatRepo, linkRepo, chatLinkRepo}
}

func (s *ChatService) RegisterChat(ctx context.Context, id int64) (*model.Chat, error) {
	chat := model.NewChat(id)
	err := s.ChatRepo.SaveChat(ctx, chat)
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, id int64) error {
	_, err := s.ChatRepo.DeleteChat(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) GetChats(ctx context.Context) ([]model.Chat, error) {
	return s.ChatRepo.GetChats(ctx)
}

func (s *ChatService) GetLinksByChatId(ctx context.Context, id int64) ([]model.Link, error) {
	links, err := s.ChatLinkRepo.GetLinksByChatID(ctx, id)
	if err != nil {
		return nil, err
	}
	return links, err
}

// TODO: save tags
func (s *ChatService) AddLink(ctx context.Context, chatId int64, req dto.AddLinkRequest) (*model.Link, error) {
	link := model.NewLink(req.Link)

	addedLink, err := s.LinkRepo.AddLink(ctx, *link)
	if err != nil {
		return nil, err
	}

	err = s.ChatLinkRepo.Subscribe(ctx, chatId, addedLink.Id)
	if err != nil {
		return nil, err
	}

	return addedLink, nil
}

func (s *ChatService) DeleteLink(ctx context.Context, chatId int64, req dto.DeleteLinkRequest) (*model.Link, error) {
	//link, err := s.linkRepo.DeleteLinkByName(chatId, req.Link)
	//if err != nil {
	//	return nil, err
	//}
	//return link, nil
	link, err := s.LinkRepo.GetLinkByName(ctx, req.Link)
	if err != nil {
		return nil, model.ErrNotFound
	}

	err = s.ChatLinkRepo.Unsubscribe(ctx, chatId, link.Id)
	if err != nil {
		return nil, err
	}
	return link, nil
}

// TODO: have id here
func (s *ChatService) UpdateLink(ctx context.Context, link *model.Link) (*model.Link, error) {
	updatedLink, err := s.LinkRepo.UpdateLink(ctx, link.Id, link)
	if err != nil {
		return nil, err
	}
	return updatedLink, nil
}

func (s *ChatService) GetLinks(ctx context.Context) ([]model.Link, error) {
	return s.LinkRepo.GetLinks(ctx)
}

func (s *ChatService) GetSubscribersByLink(ctx context.Context, link *model.Link) ([]int64, error) {
	return s.ChatLinkRepo.GetChatIdsByLinkID(ctx, link.Id)
}
