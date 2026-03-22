package model

type Tag struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func NewTag(name string) *Tag {
	return &Tag{Name: name}
}
