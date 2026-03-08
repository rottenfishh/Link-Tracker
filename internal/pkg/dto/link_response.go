package dto

import (
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkResponse struct {
	Id   string   `json:"id"`
	Link string   `json:"link"`
	Tags []string `json:"tags"`
}

func ToLinkResponse(link model.Link) *LinkResponse {
	return &LinkResponse{Id: strconv.FormatInt(link.Id, 10), Link: link.Link, Tags: link.Tags}
}
