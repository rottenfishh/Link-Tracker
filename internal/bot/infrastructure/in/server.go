package in

import (
	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type Server struct {
	Updates chan domain.LinkUpdate
	router  *gin.Engine
	url     string
}

func NewServer(url string) *Server {
	router := gin.Default()
	updChan := make(chan domain.LinkUpdate)
	handler := HttpHandler{updChan}
	registerRoutes(router, handler)
	return &Server{
		Updates: updChan,
		router:  router,
		url:     url,
	}
}

func registerRoutes(router *gin.Engine, handler HttpHandler) {
	router.POST("/updates", handler.UpdateFromLink)
}

func (server *Server) GetUpdates() chan domain.LinkUpdate {
	return server.Updates
}

func (server *Server) Run() error {
	return server.router.Run(server.url)
}
