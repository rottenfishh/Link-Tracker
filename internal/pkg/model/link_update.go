package model

type LinkUpdate struct {
	Id          int64   `json:"id"`
	Link        string  `json:"link"`
	Description string  `json:"description"`
	TgChatIds   []int64 `json:"tgChatIDs"`
}

func NewLinkUpdate(id int64, url string, description string, tgChatIds []int64) *LinkUpdate {
	return &LinkUpdate{
		Id:          id,
		Link:        url,
		Description: description,
		TgChatIds:   tgChatIds,
	}
}
