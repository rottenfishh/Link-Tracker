//nolint:wrapcheck // grpc client methods intentionally pass through transport errors to command layer
package out

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/mapper"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/scrapper"
)

type ScrapperGrpcClient struct {
	pb.ScrapperServiceClient
}

func (c *ScrapperGrpcClient) RegisterChat(ctx context.Context, chatID int64) error {
	id := mapper.ToProtoChatID(chatID)
	_, err := c.ScrapperServiceClient.RegisterChat(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (c *ScrapperGrpcClient) DeleteChat(ctx context.Context, chatID int64) error {
	id := mapper.ToProtoChatID(chatID)
	_, err := c.ScrapperServiceClient.DeleteChat(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (c *ScrapperGrpcClient) RegisterLink(ctx context.Context, chatID int64, link string, tags []string) error {
	request := dto.AddLinkRequest{
		Link: link,
		Tags: tags,
	}
	req := mapper.ToProtoAddLinkRequest(chatID, &request)
	_, err := c.ScrapperServiceClient.AddLink(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *ScrapperGrpcClient) DeleteLink(ctx context.Context, chatID int64, link string) error {
	req := mapper.ToProtoDeleteLinkRequest(chatID, link)
	_, err := c.ScrapperServiceClient.DeleteLink(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *ScrapperGrpcClient) GetLinks(ctx context.Context, chatID int64, tag string) (*dto.ListLinksResponse, error) {
	links, err := c.ScrapperServiceClient.GetLinksByChatID(ctx, mapper.ToProtoGetLinksReq(chatID, tag))
	if err != nil {
		return nil, err
	}
	linksResp := make([]dto.LinkResponse, 0)
	for _, link := range links.Links {
		linksResp = append(linksResp, *mapper.ToDomainLinkResponse(link))
	}

	resp := dto.ListLinksResponse{
		Links: linksResp,
		Size:  int32(len(linksResp)),
	}
	return &resp, nil
}
