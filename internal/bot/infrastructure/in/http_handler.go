package in

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/domain"
)

func UpdateFromLink(c *gin.Context) {
	var update domain.LinkUpdate
	if err := c.BindJSON(&update); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	// update users
	c.IndentedJSON(http.StatusOK, update)
}
