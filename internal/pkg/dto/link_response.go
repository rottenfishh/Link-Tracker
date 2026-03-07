package dto

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type LinkResponse struct {
	Id      int64    `json:"id"`
	Link    string   `json:"link"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}

func ToLinkResponse(link model.Link) *LinkResponse {
	return &LinkResponse{Id: link.Id, Link: link.Link, Tags: link.Tags}
}
