package service

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

type ScrapperService struct {
	scrapperClient ScrapperClient
}

func NewScrapperService(scrapperCLient ScrapperClient) *ScrapperService {
	return &ScrapperService{scrapperClient: scrapperCLient}
}

func (s *ScrapperService) AddLink(ctx context.Context, chatID int64, link string, tags []string) error {
	req := dto.AddLinkRequest{
		Link: link,
		Tags: tags,
	}
	err := s.scrapperClient.RegisterLink(ctx, chatID, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScrapperService) DeleteLink(ctx context.Context, chatID int64, link string) error {
	req := dto.DeleteLinkRequest{Link: link}
	err := s.scrapperClient.DeleteLink(ctx, chatID, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScrapperService) GetLinks(ctx context.Context, chatID int64) (*dto.ListLinksResponse, error) {
	links, err := s.scrapperClient.GetLinks(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (s *ScrapperService) RegisterChat(ctx context.Context, chatID int64) error {
	err := s.scrapperClient.RegisterChat(ctx, chatID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScrapperService) DeleteChat(ctx context.Context, chatID int64) error {
	err := s.scrapperClient.DeleteChat(ctx, chatID)
	if err != nil {
		return err
	}
	return nil
}
