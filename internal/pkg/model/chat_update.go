package model

type ChatUpdate struct {
	UpdateID int64
	ChatID   int64
	Message  *Message
}
