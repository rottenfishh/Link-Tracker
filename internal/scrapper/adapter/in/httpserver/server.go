package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type Server struct {
	router  *gin.Engine
	url     string
	handler *Handler
}

func NewServer(url string, service *service.ChatService) *Server {
	handler := NewHandler(service)
	server := &Server{router: gin.Default(), url: url, handler: handler}
	registerRoutes(server.router, handler)
	return server
}

func registerRoutes(router *gin.Engine, handler *Handler) {
	router.POST("/tg-chat/:id", handler.RegisterChat)
	router.DELETE("/tg-chat/:id", handler.DeleteChat)
	router.GET("/links/:id", handler.GetLinksByChatID)
	router.POST("/links/:id", handler.AddLink)
	router.DELETE("/links/:id", handler.DeleteLink)
}

func (server *Server) Run() error {
	if err := server.router.Run(server.url); err != nil {
		return fmt.Errorf("running scrapper http server on %s: %w", server.url, err)
	}
	return nil
}
