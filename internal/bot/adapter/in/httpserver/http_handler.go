package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Handler struct {
	publisher *service.UpdatePublisher
}

func NewHandler(publisher *service.UpdatePublisher) *Handler {
	return &Handler{publisher: publisher}
}

func (h *Handler) UpdateFromLink(c *gin.Context) {
	var update model.LinkUpdate
	if err := c.BindJSON(&update); err != nil {
		errResp := dto.NewRequestParsingError(err)
		slog.Error("receiving update error ", "error: ", err.Error())
		c.IndentedJSON(http.StatusBadRequest, errResp)
		return
	}
	slog.Info("received update in bot: ", "update", update)
	h.publisher.PublishUpdateNoWait(update)
	c.IndentedJSON(http.StatusOK, update)
}
