package domain

import "time"

// TODO: это будет manyTomany связь, поэтому по идее вообще субскрайберы будут в отдельной таблице, но пока так
// TODO: extract domain
type Link struct {
	Id           int64
	Link         string    `json:"link"`
	Domain       string    `json:"domain"`
	Etag         string    `json:"etag,omitempty"`
	LastModified time.Time `json:"last_modified"`
	Tags         []string  `json:"tags"`
	Events       []string  `json:"events"`
}

func NewLink(link string, tags []string) *Link {
	return &Link{Link: link, Tags: tags}
}
