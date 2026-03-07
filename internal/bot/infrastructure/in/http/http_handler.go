package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

type HttpHandler struct {
	publisher *application.UpdatePublisher
}

func NewHttpHandler(publisher *application.UpdatePublisher) *HttpHandler {
	return &HttpHandler{publisher: publisher}
}

// TODO: return only message here
func (h *HttpHandler) UpdateFromLink(c *gin.Context) {
	var update domain.LinkUpdate
	if err := c.BindJSON(&update); err != nil {
		errResp := dto.NewRequestParsingError(err)
		slog.Error("receiving update error ", "error: ", err.Error())
		c.IndentedJSON(http.StatusBadRequest, errResp)
	}
	slog.Info("received update in bot: ", "update", update)
	h.publisher.PublishUpdate(update)
	c.IndentedJSON(http.StatusOK, update)
}
