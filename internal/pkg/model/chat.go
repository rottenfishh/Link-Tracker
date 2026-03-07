package model

type Chat struct {
	Id    int64  `json:"id"`
	Links []Link `json:"links"`
}

func NewChat(id int64) *Chat {
	return &Chat{Id: id}
}
