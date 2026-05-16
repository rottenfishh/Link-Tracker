package model

import "time"

type LinkUpdate struct {
	ID        string  `json:"id"`
	Link      string  `json:"link"`
	Update    Update  `json:"update"`
	TgChatIDs []int64 `json:"tgChatIDs"`
}

type Update struct {
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	TimeCreated  time.Time `json:"timeCreated"`
	LastModified time.Time `json:"lastModified"`
	Description  string    `json:"description"`
}

func NewLinkUpdate(id string, url string, upd Update, tgChatIDs []int64) *LinkUpdate {
	return &LinkUpdate{
		ID:        id,
		Link:      url,
		Update:    upd,
		TgChatIDs: tgChatIDs,
	}
}
