package domain

type Link struct {
	Link         string   `json:"link"`
	Etag         string   `json:"etag,omitempty"`
	LastModified string   `json:"last_modified"`
	Tags         []string `json:"tags"`
	Events       []string `json:"events"`
}
