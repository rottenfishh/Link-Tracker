package domain

type Link struct {
	Link   string   `json:"link"`
	Tags   []string `json:"tags"`
	Events []string `json:"events"`
}
