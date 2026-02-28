package domain

import "time"

// TODO: это будет manyTomany связь, поэтому по идее вообще субскрайберы будут в отдельной таблице, но пока так
type Link struct {
	Link         string    `json:"link"`
	Etag         string    `json:"etag,omitempty"`
	LastModified time.Time `json:"last_modified"`
	Tags         []string  `json:"tags"`
	Events       []string  `json:"events"`
}
