package in

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/dto"
)

type HttpHandler struct {
	Updates chan domain.LinkUpdate
}

func (h *HttpHandler) UpdateFromLink(c *gin.Context) {
	var update domain.LinkUpdate
	if err := c.BindJSON(&update); err != nil {
		errResp := dto.NewRequestParsingError(err)
		slog.Error("receiving update error ", "error: ", err.Error())
		c.IndentedJSON(http.StatusBadRequest, errResp)
	}
	slog.Info("received update in bot: ", "update", update)
	h.Updates <- update
	c.IndentedJSON(http.StatusOK, update)
}
