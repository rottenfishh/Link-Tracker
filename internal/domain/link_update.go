package domain

type LinkUpdate struct {
	Id          int64   `json:"id"`
	Url         string  `json:"url"`
	Description string  `json:"description"`
	TgChatIds   []int64 `json:"tgChatIds"`
}

func NewLinkUpdate(id int64, url string, description string, tgChatIds []int64) *LinkUpdate {
	return &LinkUpdate{
		Id:          id,
		Url:         url,
		Description: description,
		TgChatIds:   tgChatIds,
	}
}
