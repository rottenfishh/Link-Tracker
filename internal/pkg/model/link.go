package model

import (
	"strings"
	"time"
)

// TODO: это будет manyTomany связь, поэтому по идее вообще субскрайберы будут в отдельной таблице, но пока так
// TODO: extract model
type Link struct {
	Id          int64
	Link        string    `json:"url"`
	Domain      string    `json:"model"`
	LastUpdated time.Time `json:"last_updated"`
}

// TODO: print message to user if no updater for this link
func NewLink(link string) *Link {
	domain := "unknown"
	if strings.Contains(link, "github") {
		domain = "github"
	}
	if strings.Contains(link, "stackoverflow") {
		domain = "stackof"
	}
	return &Link{Link: link, Domain: domain, LastUpdated: time.Now()}
}
