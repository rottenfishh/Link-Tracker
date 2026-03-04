package in

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/application"
)

// TODO: create service layer
// TODO: кидать свои собственные ошибки
type HttpHandler struct {
	service *application.ChatService
}

func NewHttpHandler(service *application.ChatService) *HttpHandler {
	return &HttpHandler{service: service}
}

func (h *HttpHandler) RegisterChat(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		errResp := dto.NewRequestParsingError(err)
		c.JSON(http.StatusBadRequest, errResp)
	}

	chat, err := h.service.RegisterChat(idInt)
	if err != nil {
		errResp := dto.NewRepositoryError("Error saving chat", err)
		c.IndentedJSON(http.StatusInternalServerError, errResp)
	}

	c.IndentedJSON(http.StatusOK, *chat)
}

func (h *HttpHandler) DeleteChat(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		errResp := dto.NewRequestParsingError(err)
		c.JSON(http.StatusBadRequest, errResp)
	}

	err = h.service.DeleteChat(idInt)
	if err != nil {
		errResp := dto.NewRepositoryError("Error deleting chat", err)
		c.JSON(http.StatusInternalServerError, errResp)
	}
	c.IndentedJSON(http.StatusOK, gin.H{"id": id})
}

func (h *HttpHandler) GetLinksByChatId(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		errResp := dto.NewRequestParsingError(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	links, err := h.service.GetLinksByChatId(idInt)
	if err != nil {
		errResp := dto.NewRepositoryError("Error getting links", err)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	linkResp := make([]dto.LinkResponse, len(links))
	for _, link := range links {
		linkResp = append(linkResp, *dto.ToLinkResponse(link))
	}

	c.IndentedJSON(http.StatusOK, linkResp)
}

func (h *HttpHandler) AddLink(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		errResp := dto.NewRequestParsingError(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}
	var link dto.AddLinkRequest
	if err := c.BindJSON(&link); err != nil {
		errResp := dto.NewRequestParsingError(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	addedLink, err := h.service.AddLink(idInt, link)
	if err != nil {
		errResp := dto.NewRepositoryError("Error adding link", err)
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}
	linkResp := dto.ToLinkResponse(*addedLink)

	c.IndentedJSON(http.StatusOK, linkResp)
}

func (h *HttpHandler) DeleteLink(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		errResp := dto.NewRequestParsingError(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	var link dto.DeleteLinkRequest
	if err := c.BindJSON(&link); err != nil {
		errResp := dto.NewRequestParsingError(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	err = h.service.DeleteLink(idInt, link)
	if err != nil {
		errResp := dto.NewRepositoryError("Error deleting link", err)
		c.JSON(http.StatusInternalServerError, errResp)
		return
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
