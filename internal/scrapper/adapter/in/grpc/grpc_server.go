//nolint:mnd // shutdown timeout is intentionally kept inline for transport setup
package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/mapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/pkg/proto/scrapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type ScrapperServer struct {
	pb.UnimplementedScrapperServiceServer
	chatService *service.ChatService
	rateLimitMW Limiter
	port        string
	gatewayPort string
}

type Limiter interface {
	Limit(next http.Handler) http.Handler
}

func NewScrapperServer(chatService *service.ChatService, rateLimit Limiter, port, gatewayPort string) *ScrapperServer {
	return &ScrapperServer{
		chatService: chatService,
		rateLimitMW: rateLimit,
		port:        port,
		gatewayPort: gatewayPort,
	}
}

func (s *ScrapperServer) RegisterChat(ctx context.Context, chatID *pb.ChatID) (*pb.ChatResponse, error) {
	if chatID == nil || chatID.Id == 0 {
		return nil, fmt.Errorf("invalid register chat request: %w", status.Error(
			codes.InvalidArgument,
			"invalid register chat request",
		))
	}
	_, err := s.chatService.RegisterChat(ctx, chatID.Id)
	if err != nil {
		// code := parseServerCode(err)
		// message := "Error saving chat"
		// errResp := dto.NewServiceError(message, err, code)
		// c.IndentedJSON(http.StatusInternalServerError, errResp)
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to register chat: %v", err)
	}
	return &pb.ChatResponse{Message: "Chat registered"}, nil
}

func (s *ScrapperServer) DeleteChat(ctx context.Context, chatID *pb.ChatID) (*pb.ChatResponse, error) {
	err := s.chatService.DeleteChat(ctx, chatID.Id)
	if err != nil {
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to delete chat: %v", err)
	}
	return &pb.ChatResponse{Message: "Chat successfully deleted"}, nil
}

func (s *ScrapperServer) GetLinksByChatID(ctx context.Context, req *pb.GetLinksReq) (*pb.ListLinkResponse, error) {
	links, err := s.chatService.GetLinksByChatIDAndTag(ctx, req.Id, req.Tag)
	if err != nil {
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to get links: %v", err)
	}
	var linksResp pb.ListLinkResponse
	linksResp.Links = make([]*pb.LinkResponse, 0)

	for _, link := range links {
		linksResp.Links = append(linksResp.Links, mapper.ToProtoLinkResponse(&link))
	}
	linksResp.Size = int32(len(linksResp.Links))
	return &linksResp, nil
}

func (s *ScrapperServer) AddLink(ctx context.Context, req *pb.AddLinkRequest) (*pb.LinkResponse, error) {
	request := mapper.ToDomainAddLinkRequest(req)
	link, err := s.chatService.AddLink(ctx, req.ChatID, request)
	if err != nil {
		code := parseServerCode(err)
		slog.Error("add link failed", "error", err, "code", code)
		return nil, status.Errorf(code, "failed to add link: %v", err)
	}
	resp := mapper.ToProtoLinkResponse(link)
	return resp, nil
}

func (s *ScrapperServer) DeleteLink(ctx context.Context, req *pb.RemoveLinkRequest) (*pb.LinkResponse, error) {
	link, err := s.chatService.Unsubscribe(ctx, req.ChatID, mapper.ToDomainRemoveLinkRequest(req))
	if err != nil {
		code := parseServerCode(err)
		slog.Error("delete link failed", "error", err, "code", code)
		return nil, status.Errorf(code, "failed to delete link: %v", err)
	}
	resp := mapper.ToProtoLinkResponse(link)
	return resp, nil
}

func (s *ScrapperServer) RunServer(ctx context.Context) error {
	grpcServer := grpc.NewServer()
	pb.RegisterScrapperServiceServer(grpcServer, s)

	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("scrapper server error listening on port %s: %w", s.port, err)
	}
	go func() {
		slog.Info("scrapper server listening on port", "port", s.port)
		serveErr := grpcServer.Serve(lis)
		if serveErr != nil {
			slog.Error("scrapper server error serving on port", "port", s.port, "error", serveErr)
		}
	}()

	gatewayServer, err := buildGatewayClient(s.port, s.gatewayPort, s.rateLimitMW)
	if err != nil {
		return err
	}
	go func() {
		slog.Info("Serving gRPC-Gateway on", "port", s.gatewayPort)
		serveErr := gatewayServer.ListenAndServe()
		if serveErr != nil {
			slog.Error("failed to serve gRPC-Gateway", "error", serveErr)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down scrapper grpc server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if shutdownErr := gatewayServer.Shutdown(shutdownCtx); shutdownErr != nil {
		slog.Error("failed to shutdown gRPC-Gateway", "error", shutdownErr)
	}
	grpcServer.GracefulStop()
	return nil
}

func buildGatewayClient(serverPort string, gatewayPort string, rateLimitMW Limiter) (*http.Server, error) {
	conn, err := grpc.NewClient(
		":"+serverPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	mux := runtime.NewServeMux()

	muxWithLimiter := rateLimitMW.Limit(mux)

	err = pb.RegisterScrapperServiceHandler(context.Background(), mux, conn)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("failed to close scrapper grpc gateway connection", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	gwServer := &http.Server{
		Addr:    ":" + gatewayPort,
		Handler: muxWithLimiter,
	}
	return gwServer, nil
}

func parseServerCode(err error) codes.Code {
	var code codes.Code
	switch {
	case errors.Is(err, model.ErrNotFound):
		code = codes.NotFound
	case errors.Is(err, model.ErrLinkAlreadyTracked):
		code = codes.AlreadyExists
	case errors.Is(err, model.ErrInvalidRequest):
		code = codes.InvalidArgument
	default:
		code = codes.Internal
	}
	return code
}
