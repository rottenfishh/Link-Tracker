package application

import (
	"context"
)

type Command interface {
	Name() string
	Description() string
	Execute(ctx *context.Context, state *State) (*CommandResult, error)
}

type Message struct {
	Text string
	/// smth later
}

func NewMessage(text string) *Message {
	return &Message{text}
}

type CommandResult struct {
	IsFinished bool
	Message    *Message
}

func NewCommandResult(isFinished bool, message string) *CommandResult {
	return &CommandResult{isFinished, &Message{message}}
}
