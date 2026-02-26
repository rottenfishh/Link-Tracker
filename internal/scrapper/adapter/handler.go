package adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/domain"
)

func RegisterChat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id is required"})
	}
	chat := domain.NewChat(id)
	//repo.save(chat)

	c.IndentedJSON(http.StatusOK, chat)
}

func DeleteChat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id is required"})
	}
	//repo.delete(id)
	c.IndentedJSON(http.StatusOK, gin.H{"id": id})
}

func GetLinksByChatId(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id is required"})
	}
	//repo.getLinks(id)
	links := make([]domain.Link, 0)
	c.IndentedJSON(http.StatusOK, links)
}

func AddLink(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id is required"})
	}
	var link domain.Link
	if err := c.BindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	// app.TrackLink(id, link)
	c.IndentedJSON(http.StatusOK, link)
}

func DeleteLink(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id is required"})
	}
	var link domain.Link
	if err := c.BindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	// app.deleteLink(link.link)
	c.IndentedJSON(http.StatusOK, gin.H{"id": id})
}
