package in

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/mapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/bot"
	"google.golang.org/grpc"
)

type BotServer struct {
	bot.UnimplementedBotServiceServer
	publisher application.UpdatePublisher
}

func NewBotServiceServer(publisher application.UpdatePublisher) *BotServer {
	return &BotServer{publisher: publisher}
}

func (s *BotServer) SendUpdate(ctx context.Context, update *bot.LinkUpdate) (*bot.UpdateResponse, error) {
	upd := mapper.ToDomainLinkUpdate(update)
	slog.Info("Received updated in bot: ", "update", upd)
	s.publisher.PublishUpdate(*upd)
	return &bot.UpdateResponse{Message: "Update received"}, nil
}

// TODO: port from config
func RunGRPCServer(port string, publisher application.UpdatePublisher) error {
	botServer := NewBotServiceServer(publisher)
	grpcServer := grpc.NewServer()
	bot.RegisterBotServiceServer(grpcServer, botServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		slog.Error("bot server error listening for grpc on port ", "port", port)
		return err
	}
	err = grpcServer.Serve(lis)
	if err != nil {
		slog.Error("error serving grpc server on port ", "port", port)
		return err
	}
	return nil
}
