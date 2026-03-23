package model

type Chat struct {
	ChatID int64 `json:"chat_id"`
	UserID int64 `json:"user_id"`
}

func NewChat(id int64) *Chat {
	return &Chat{ChatID: id}
}
