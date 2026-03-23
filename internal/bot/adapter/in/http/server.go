package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Server struct {
	router    *gin.Engine
	publisher *service.UpdatePublisher
	url       string
}

func NewServer(url string, publisher *service.UpdatePublisher) *Server {
	router := gin.Default()
	handler := NewHttpHandler(publisher)
	registerRoutes(router, handler)
	return &Server{
		router:    router,
		url:       url,
		publisher: publisher,
	}
}

func registerRoutes(router *gin.Engine, handler *HttpHandler) {
	router.POST("/updates", handler.UpdateFromLink)
}

func (server *Server) GetUpdates() chan model.LinkUpdate {
	return server.publisher.GetUpdates()
}

func (server *Server) Run() error {
	return server.router.Run(server.url)
}
