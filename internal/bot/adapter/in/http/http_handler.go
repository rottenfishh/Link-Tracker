package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type HttpHandler struct {
	publisher *service.UpdatePublisher
}

func NewHttpHandler(publisher *service.UpdatePublisher) *HttpHandler {
	return &HttpHandler{publisher: publisher}
}

// TODO: return only message here
func (h *HttpHandler) UpdateFromLink(c *gin.Context) {
	var update model.LinkUpdate
	if err := c.BindJSON(&update); err != nil {
		errResp := dto.NewRequestParsingError(err)
		slog.Error("receiving update error ", "error: ", err.Error())
		c.IndentedJSON(http.StatusBadRequest, errResp)
	}
	slog.Info("received update in bot: ", "update", update)
	h.publisher.PublishUpdate(update)
	c.IndentedJSON(http.StatusOK, update)
}
