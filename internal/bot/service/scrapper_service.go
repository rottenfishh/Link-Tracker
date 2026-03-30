package service

//
//import (
//	"context"
//
//    "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/commands"
//    "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
//)
//
//type ScrapperClient struct {
//	scrapperClient commands.ScrapperClient
//}
//
//func NewScrapperService(scrapperCLient commands.ScrapperClient) *ScrapperClient {
//	return &ScrapperClient{scrapperClient: scrapperCLient}
//}
//
//func (s *ScrapperClient) AddLink(ctx context.Context, chatID int64, link string, tags []string) error {
//	req := dto.AddLinkRequest{
//		Link: link,
//		Tags: tags,
//	}
//	err := s.scrapperClient.RegisterLink(ctx, chatID, req)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *ScrapperClient) DeleteLink(ctx context.Context, chatID int64, link string) error {
//	req := dto.DeleteLinkRequest{Link: link}
//	err := s.scrapperClient.DeleteLink(ctx, chatID, req)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *ScrapperClient) GetLinks(ctx context.Context, chatID int64, tag string) (*dto.ListLinksResponse, error) {
//	links, err := s.scrapperClient.GetLinks(ctx, chatID, tag)
//	if err != nil {
//		return nil, err
//	}
//	return links, nil
//}
//
//func (s *ScrapperClient) RegisterChat(ctx context.Context, chatID int64) error {
//	err := s.scrapperClient.RegisterChat(ctx, chatID)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *ScrapperClient) DeleteChat(ctx context.Context, chatID int64) error {
//	err := s.scrapperClient.DeleteChat(ctx, chatID)
//	if err != nil {
//		return err
//	}
//	return nil
//}
