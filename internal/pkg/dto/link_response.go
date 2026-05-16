package dto

import (
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkResponse struct {
	ID   string   `json:"id"`
	Link string   `json:"link"`
	Tags []string `json:"tags"`
}

// TODO: get tags
func ToLinkResponse(link model.Link) *LinkResponse {
	return &LinkResponse{ID: strconv.FormatInt(link.ID, 10), Link: link.Link, Tags: link.Tags}
}
