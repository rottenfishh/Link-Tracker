package commands

import "context"

type Command interface {
	Name() string
	Description() string
	Execute(ctx *context.Context, args []string) (*Message, error)
}

type Message struct {
	Text string
	/// smth later
}

func NewMessage(text string) *Message {
	return &Message{text}
}
