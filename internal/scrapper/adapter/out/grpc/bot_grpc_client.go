package grpc

import (
	"context"
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/mapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/bot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BotGrpcClient struct {
	pb.BotServiceClient
}

func NewBotGrpcClient(port string) (*BotGrpcClient, error) {
	conn, err := grpc.NewClient(":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to port %s: %w", port, err)
	}
	return &BotGrpcClient{pb.NewBotServiceClient(conn)}, nil
}

// TODO: implement
func (c *BotGrpcClient) SendReport(ctx context.Context, report model.Report) error {
	rep := mapper.ToProtoReport(&report)
	_, err := c.BotServiceClient.SendReport(ctx, rep)
	if err != nil {
		return fmt.Errorf("sending report to bot grpc service: %w", err)
	}
	return nil
}

func (c *BotGrpcClient) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	upd := mapper.ToProtoLinkUpdate(&update)
	_, err := c.BotServiceClient.SendUpdate(ctx, upd)
	if err != nil {
		return fmt.Errorf("sending update to bot grpc service: %w", err)
	}
	return nil
}
