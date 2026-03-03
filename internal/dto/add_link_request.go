package dto

type AddLinkRequest struct {
	Link    string   `json:"link"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}
