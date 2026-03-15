package service

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, state *State) (*CommandResult, error)
}

type CommandResult struct {
	IsFinished bool
	Message    *model.Message
}

func NewCommandResult(isFinished bool, message string) *CommandResult {
	return &CommandResult{isFinished, &model.Message{message}}
}
