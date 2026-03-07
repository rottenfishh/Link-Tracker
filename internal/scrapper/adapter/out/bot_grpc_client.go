package out

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
		return nil, fmt.Errorf("failed to connect to port %s %v", port, err)
	}
	return &BotGrpcClient{pb.NewBotServiceClient(conn)}, nil
}

func (c *BotGrpcClient) SendUpdate(ctx context.Context, update model.LinkUpdate) error {
	upd := mapper.ToProtoLinkUpdate(&update)
	_, err := c.BotServiceClient.SendUpdate(context.Background(), upd)
	if err != nil {
		return err
	}
	return nil
}
