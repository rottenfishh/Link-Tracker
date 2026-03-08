package in

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/mapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/bot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type BotServer struct {
	bot.UnimplementedBotServiceServer
	publisher *service.UpdatePublisher
}

func NewBotServiceServer(publisher *service.UpdatePublisher) *BotServer {
	return &BotServer{publisher: publisher}
}

func (s *BotServer) SendUpdate(ctx context.Context, update *bot.LinkUpdate) (*bot.UpdateResponse, error) {
	if update == nil || update.Link == "" {
		slog.Error("code", codes.InvalidArgument.String(), "message", "Update is empty")
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid update request",
		)
	}

	upd := mapper.ToDomainLinkUpdate(update)
	slog.Info("Received updated in bot: ", "update", upd)
	s.publisher.PublishUpdate(*upd)
	return &bot.UpdateResponse{Message: "Update received"}, nil
}

// TODO: port from config
func (s *BotServer) RunServer(port string, gatewayPort string) error {
	grpcServer := grpc.NewServer()
	bot.RegisterBotServiceServer(grpcServer, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		slog.Error("bot server error listening for grpc on port ", "port", port)
		return err
	}

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			slog.Error("error serving grpc server on port ", "port", port)
		}
	}()

	err = runClient(port, gatewayPort)
	if err != nil {
		return err
	}
	return nil
}

func runClient(serverPort string, gatewayPort string) error {
	conn, err := grpc.NewClient(
		":"+serverPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}

	mux := runtime.NewServeMux()

	err = bot.RegisterBotServiceHandler(context.Background(), mux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	gwServer := &http.Server{
		Addr:    ":" + gatewayPort,
		Handler: mux,
	}

	slog.Info("Serving gRPC-Gateway on", "port", gatewayPort)
	err = gwServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to serve gRPC-Gateway: %w", err)
	}
	return nil
}

func (s *BotServer) GetUpdates() chan model.LinkUpdate {
	return s.publisher.GetUpdates()
}
