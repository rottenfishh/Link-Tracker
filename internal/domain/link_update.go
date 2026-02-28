package domain

type LinkUpdate struct {
	Id          int64   `json:"id"`
	Url         string  `json:"url"`
	Description string  `json:"description"`
	TgChatIds   []int64 `json:"tgChatIds"`
}
