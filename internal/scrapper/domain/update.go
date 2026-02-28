package domain

import "time"

type Update struct {
	LastModified time.Time
	Description  string
}

func NewUpdate(time time.Time, description string) *Update {
	return &Update{time, description}
}
