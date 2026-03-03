package in

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure/out"
)

// TODO: use goddamn dtos please
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

	chat := domain.NewChat(idInt)
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
		return
	}

	links, err := h.repo.GetLinksById(idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	var link dto.AddLinkRequest
	if err := c.BindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	linkDomain := domain.NewLink(link.Link, link.Tags)

	addedLink, err := h.repo.AddLink(idInt, *linkDomain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	linkResp := dto.ToLinkResponse(addedLink)

	c.IndentedJSON(http.StatusOK, linkResp)
}

// TODO: accept only link in body
func (h *HttpHandler) DeleteLink(c *gin.Context) {
	id := c.Param("id")

	idInt, err := parseId(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var link dto.AddLinkRequest
	if err := c.BindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.repo.DeleteLink(idInt, link.Link)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
