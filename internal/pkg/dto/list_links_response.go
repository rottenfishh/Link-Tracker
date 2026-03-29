package dto

type ListLinksResponse struct {
	Links []LinkResponse `json:"links"`
	Size  int            `json:"size"`
}
