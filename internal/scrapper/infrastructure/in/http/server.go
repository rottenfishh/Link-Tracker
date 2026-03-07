package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/application"
)

type Server struct {
	router  *gin.Engine
	url     string
	handler *HttpHandler
}

func NewServer(url string, service *application.ChatService) *Server {
	handler := NewHttpHandler(service)
	server := &Server{router: gin.Default(), url: url, handler: handler}
	registerRoutes(server.router, handler)
	return server
}

func registerRoutes(router *gin.Engine, handler *HttpHandler) {
	router.POST("/tg-chat/:id", handler.RegisterChat)
	router.DELETE("/tg-chat/:id", handler.DeleteChat)
	router.GET("/links/:id", handler.GetLinksByChatId)
	router.POST("/links/:id", handler.AddLink)
	router.DELETE("/links/:id", handler.DeleteLink)
}

func (server *Server) Run() error {
	return server.router.Run(server.url)
}
