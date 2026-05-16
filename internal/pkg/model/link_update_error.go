package model

type LinkUpdateError struct {
	Link    string `json:"link"`
	Err     string `json:"err"`
	Message string `json:"message"`
	LinkID  int64  `json:"link_id"`
}
