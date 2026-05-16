package model

type Report struct {
	ID               string             `json:"id"`
	ChatID           int64              `json:"chat_id"`
	LinkUpdateErrors []*LinkUpdateError `json:"errors"`
}

func NewReport(id string, chatID int64) *Report {
	errors := make([]*LinkUpdateError, 0)
	return &Report{ID: id, ChatID: chatID, LinkUpdateErrors: errors}
}
