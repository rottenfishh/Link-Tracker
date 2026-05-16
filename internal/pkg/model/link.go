package model

import (
	"strings"
	"time"
)

type Link struct {
	ID            int64     `json:"id"`
	Link          string    `json:"url"`
	Title         string    `json:"title"`
	FormattedLink string    `json:"formatted_link"`
	Domain        string    `json:"model"`
	LastUpdated   time.Time `json:"last_updated"`
	Tags          []string  `json:"tags"`
}

// TODO: print message to user if no updater for this link
func NewLink(link string, tags []string) *Link {
	domain := "unknown"
	if strings.Contains(link, "github") {
		domain = "github"
	}
	if strings.Contains(link, "stackoverflow") {
		domain = "stackof"
	}
	return &Link{Link: link, Tags: tags, Domain: domain}
}
