package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Server struct {
	Updates chan model.LinkUpdate
	router  *gin.Engine
	url     string
}

func NewServer(url string, publisher *service.UpdatePublisher) *Server {
	router := gin.Default()
	updChan := make(chan model.LinkUpdate)
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

func (server *Server) GetUpdates() chan model.LinkUpdate {
	return server.Updates
}

func (server *Server) Run() error {
	return server.router.Run(server.url)
}
