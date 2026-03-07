package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

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
}

func NewScrapperServer(chatService *service.ChatService) *ScrapperServer {
	return &ScrapperServer{chatService: chatService}
}

func (s *ScrapperServer) RegisterChat(ctx context.Context, chatID *pb.ChatID) (*pb.ChatResponse, error) {
	if chatID == nil || chatID.Id == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid register chat request",
		)
	}
	_, err := s.chatService.RegisterChat(chatID.Id)
	if err != nil {
		//code := parseServerCode(err)
		//		message := "Error saving chat"
		//		errResp := dto.NewServiceError(message, err, code)
		//		c.IndentedJSON(http.StatusInternalServerError, errResp)
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to register chat: %v", err)
	}
	return &pb.ChatResponse{Message: "Chat registered"}, nil
}

func (s *ScrapperServer) DeleteChat(ctx context.Context, chatID *pb.ChatID) (*pb.ChatResponse, error) {
	err := s.chatService.DeleteChat(chatID.Id)
	if err != nil {
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to delete chat: %v", err)
	}
	return &pb.ChatResponse{Message: "Chat successfully deleted"}, nil
}

func (s *ScrapperServer) GetLinksByChatID(ctx context.Context, chatID *pb.ChatID) (*pb.ListLinkResponse, error) {
	links, err := s.chatService.GetLinksByChatId(chatID.Id)
	if err != nil {
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to get links: %v", err)
	}
	var linksResp pb.ListLinkResponse
	linksResp.Links = make([]*pb.LinkResponse, 0)

	for _, link := range links {
		linksResp.Links = append(linksResp.Links, mapper.ToProtoLinkResponse(&link))
	}
	return &linksResp, nil
}

func (s *ScrapperServer) AddLink(ctx context.Context, req *pb.AddLinkRequest) (*pb.LinkResponse, error) {
	request := mapper.ToDomainAddLinkRequest(req)
	link, err := s.chatService.AddLink(req.ChatID, request)
	if err != nil {
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to add link: %v", err)
	}
	resp := mapper.ToProtoLinkResponse(link)
	return resp, nil
}

func (s *ScrapperServer) DeleteLink(ctx context.Context, req *pb.RemoveLinkRequest) (*pb.LinkResponse, error) {
	link, err := s.chatService.DeleteLink(req.ChatID, mapper.ToDomainRemoveLinkRequest(req))
	if err != nil {
		code := parseServerCode(err)
		return nil, status.Errorf(code, "failed to delete link: %v", err)
	}
	resp := mapper.ToProtoLinkResponse(link)
	return resp, nil
}

func (s *ScrapperServer) RunServer(port string, gatewayPort string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("scrapper server error listening on port %s", port)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterScrapperServiceServer(grpcServer, s)

	go func() {
		err = grpcServer.Serve(lis)

		if err != nil {
			slog.Error("scrapper server error serving on port", "port", port)
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

	err = pb.RegisterScrapperServiceHandler(context.Background(), mux, conn)
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
