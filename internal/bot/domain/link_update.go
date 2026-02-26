package domain

type LinkUpdate struct {
	Id          string   `json:"id"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	TgChatIds   []string `json:"tgChatIds"`
}
