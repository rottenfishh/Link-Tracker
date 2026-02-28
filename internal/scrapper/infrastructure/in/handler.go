package in

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	domain2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure/out"
)

type HttpHandler struct {
	repo out.ChatRepository
}

func NewHttpHandler(repo out.ChatRepository) *HttpHandler {
	return &HttpHandler{repo}
}

func (h *HttpHandler) RegisterChat(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	chat := domain2.NewChat(idInt)
	err = h.repo.SaveChat(chat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	c.IndentedJSON(http.StatusOK, chat)
}

func (h *HttpHandler) DeleteChat(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	err = h.repo.DeleteChat(idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"id": id})
}

func (h *HttpHandler) GetLinksByChatId(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	links, err := h.repo.GetLinksById(idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, links)
}

func (h *HttpHandler) AddLink(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	var link domain2.Link
	if err := c.BindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	addedLink, err := h.repo.AddLink(idInt, link)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, addedLink)
}

// TODO: accept only link in body
func (h *HttpHandler) DeleteLink(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	var link domain2.Link
	if err := c.BindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	err = h.repo.DeleteLink(idInt, link.Link)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"id": id})
}

func parseId(id string) (int64, error) {
	if id == "" {
		return -1, fmt.Errorf("id is required")
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id format. should be number")
	}
	return idInt, nil
}
