//nolint:mnd // small configuration-like constants keep the transport code readable here
package in

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/mapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/bot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type BotServer struct {
	bot.UnimplementedBotServiceServer
	publisher   *service.UpdatePublisher
	port        string
	gatewayPort string
}

func NewBotServiceServer(publisher *service.UpdatePublisher, port, gatewayPort string) *BotServer {
	return &BotServer{publisher: publisher, port: port, gatewayPort: gatewayPort}
}

func (s *BotServer) SendUpdate(ctx context.Context, update *bot.LinkUpdate) (*bot.UpdateResponse, error) {
	_ = ctx
	if update == nil || update.Link == "" {
		slog.Error("invalid update request", "code", codes.InvalidArgument, "message", "Update is empty")
		return nil, fmt.Errorf("invalid update request: %w", status.Error(
			codes.InvalidArgument,
			"invalid update request",
		))
	}

	upd := mapper.ToDomainLinkUpdate(update)
	slog.Info("Received updated in bot: ", "update", update)

	s.publisher.PublishUpdateNoWait(*upd)
	return &bot.UpdateResponse{Message: "Update received"}, nil
}

func (s *BotServer) SendReport(ctx context.Context, report *bot.Report) (*bot.UpdateResponse, error) {
	_ = ctx
	if report == nil {
		slog.Error("invalid report request", "code", codes.InvalidArgument, "message", "Report is empty")
		return nil, fmt.Errorf("invalid report request: %w", status.Error(
			codes.InvalidArgument,
			"invalid update request"))
	}

	result := mapper.ToDomainReport(report)
	slog.Info("Received report in bot: ", "result", result)

	s.publisher.PublishReportNoWait(result)
	return &bot.UpdateResponse{Message: "Report received"}, nil
}

// TODO: port from config
func (s *BotServer) Run(ctx context.Context) error {
	grpcServer := grpc.NewServer()
	bot.RegisterBotServiceServer(grpcServer, s)

	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		slog.Error("bot server error listening for grpc on port ", "port", s.port)
		return fmt.Errorf("listening on bot grpc port %s: %w", s.port, err)
	}
	go func() {
		serveErr := grpcServer.Serve(lis)
		if serveErr != nil {
			slog.Error("error serving grpc server on port ", "port", s.port, "error", serveErr)
		}
	}()

	gatewayClient, err := buildGatewayClient(s.port, s.gatewayPort)
	if err != nil {
		return err
	}
	go func() {
		slog.Info("Serving gRPC-Gateway on", "port", s.gatewayPort)
		serveErr := gatewayClient.ListenAndServe()
		if serveErr != nil {
			slog.Error("failed to serve gRPC-Gateway", "error", serveErr)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down bot grpc server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if shutdownErr := gatewayClient.Shutdown(shutdownCtx); shutdownErr != nil {
		slog.Error("failed to shutdown gRPC-Gateway", "error", shutdownErr)
	}
	grpcServer.GracefulStop()

	return nil
}

func buildGatewayClient(serverPort string, gatewayPort string) (*http.Server, error) {
	conn, err := grpc.NewClient(
		":"+serverPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	mux := runtime.NewServeMux()

	err = bot.RegisterBotServiceHandler(context.Background(), mux, conn)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("failed to close bot grpc gateway connection", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	gwServer := &http.Server{
		Addr:    ":" + gatewayPort,
		Handler: mux,
	}

	return gwServer, nil
}

func (s *BotServer) GetUpdates() chan service.UpdateJob {
	return s.publisher.GetUpdatesChan()
}

func (s *BotServer) GetReports() chan service.ReportJob {
	return s.publisher.GetReportsChan()
}
