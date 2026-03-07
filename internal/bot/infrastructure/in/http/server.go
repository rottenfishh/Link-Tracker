package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/domain"
)

type Server struct {
	Updates chan domain.LinkUpdate
	router  *gin.Engine
	url     string
}

func NewServer(url string, publisher *application.UpdatePublisher) *Server {
	router := gin.Default()
	updChan := make(chan domain.LinkUpdate)
	handler := NewHttpHandler(publisher)
	registerRoutes(router, handler)
	return &Server{
		Updates: updChan,
		router:  router,
		url:     url,
	}
}

func registerRoutes(router *gin.Engine, handler *HttpHandler) {
	router.POST("/updates", handler.UpdateFromLink)
}

func (server *Server) GetUpdates() chan domain.LinkUpdate {
	return server.Updates
}

func (server *Server) Run() error {
	return server.router.Run(server.url)
}
