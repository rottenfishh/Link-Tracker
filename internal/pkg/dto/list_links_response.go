package dto

type ListLinksResponse struct {
	Links []LinkResponse `json:"links"`
	Size  int32          `json:"size"`
}
