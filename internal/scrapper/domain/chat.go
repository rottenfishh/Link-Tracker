package domain

type Chat struct {
	Id    string `json:"id"`
	Links []Link `json:"links"`
}

func NewChat(id string) *Chat {
	return &Chat{Id: id}
}
