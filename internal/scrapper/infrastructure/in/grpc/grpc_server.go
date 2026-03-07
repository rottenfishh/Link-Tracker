package grpc

import (
	"context"
	"fmt"
	"net"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/mapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/application"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/scrapper"
	"google.golang.org/grpc"
)

type ScrapperServer struct {
	pb.UnimplementedScrapperServiceServer
	chatService *application.ChatService
}

func NewScrapperServer(chatService *application.ChatService) *ScrapperServer {
	return &ScrapperServer{chatService: chatService}
}

func (s *ScrapperServer) RegisterChat(ctx context.Context, chatID *pb.ChatID) (*pb.ChatResponse, error) {
	_, err := s.chatService.RegisterChat(chatID.Id)
	if err != nil {
		return nil, err
	}
	return &pb.ChatResponse{Message: "Chat registered"}, nil
}

func (s *ScrapperServer) DeleteChat(ctx context.Context, chatID *pb.ChatID) (*pb.ChatResponse, error) {
	err := s.chatService.DeleteChat(chatID.Id)
	if err != nil {
		return nil, err
	}
	return &pb.ChatResponse{Message: "Chat successfully deleted"}, nil
}

func (s *ScrapperServer) GetLinksByChatID(ctx context.Context, chatID *pb.ChatID) (*pb.ListLinkResponse, error) {
	links, err := s.chatService.GetLinksByChatId(chatID.Id)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	resp := mapper.ToProtoLinkResponse(link)
	return resp, nil
}

func (s *ScrapperServer) DeleteLink(ctx context.Context, req *pb.RemoveLinkRequest) (*pb.LinkResponse, error) {
	link, err := s.chatService.DeleteLink(req.ChatID, mapper.ToDomainRemoveLinkRequest(req))
	if err != nil {
		return nil, err
	}
	resp := mapper.ToProtoLinkResponse(link)
	return resp, nil
}

func RunServer(port string, service *application.ChatService) error {
	scrapperServer := NewScrapperServer(service)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("scrapper server error listening on port %s", port)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterScrapperServiceServer(grpcServer, scrapperServer)

	err = grpcServer.Serve(lis)

	if err != nil {
		return fmt.Errorf("scrapper server error serving on port %s", port)
	}
	return nil
}
