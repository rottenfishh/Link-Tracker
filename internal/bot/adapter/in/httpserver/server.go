package httpserver

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
)

type Server struct {
	router    *gin.Engine
	publisher *service.UpdatePublisher
	url       string
}

func NewServer(url string, publisher *service.UpdatePublisher) *Server {
	router := gin.Default()
	handler := NewHandler(publisher)
	registerRoutes(router, handler)
	return &Server{
		router:    router,
		url:       url,
		publisher: publisher,
	}
}

func registerRoutes(router *gin.Engine, handler *Handler) {
	router.POST("/updates", handler.UpdateFromLink)
}

func (s *Server) GetUpdates() chan service.UpdateJob {
	return s.publisher.GetUpdatesChan()
}

func (s *Server) GetReports() chan service.ReportJob {
	return s.publisher.GetReportsChan()
}

func (s *Server) Run(_ context.Context) error {
	if err := s.router.Run(s.url); err != nil {
		return fmt.Errorf("running bot http s on %s: %w", s.url, err)
	}
	return nil
}
