package application

import (
	dto2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

type ScrapperService struct {
	scrapperClient ScrapperClient
}

func NewScrapperService(scrapperCLient ScrapperClient) *ScrapperService {
	return &ScrapperService{scrapperClient: scrapperCLient}
}

func (s *ScrapperService) AddLink(chatID int64, link string, tags []string) error {
	req := dto2.AddLinkRequest{
		Link: link,
		Tags: tags,
	}
	err := s.scrapperClient.RegisterLink(chatID, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScrapperService) DeleteLink(chatID int64, link string) error {
	req := dto2.DeleteLinkRequest{Link: link}
	err := s.scrapperClient.DeleteLink(chatID, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScrapperService) GetLinks(chatID int64) ([]dto2.LinkResponse, error) {
	links, err := s.scrapperClient.GetLinks(chatID)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (s *ScrapperService) RegisterChat(chatID int64) error {
	err := s.scrapperClient.RegisterChat(chatID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScrapperService) DeleteChat(chatID int64) error {
	err := s.scrapperClient.DeleteChat(chatID)
	if err != nil {
		return err
	}
	return nil
}
