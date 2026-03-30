package commands

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"

type CommandResult struct {
	Finished bool
	Message  *model.Message
}

func NewCommandResult(isFinished bool, message string) *CommandResult {
	return &CommandResult{isFinished, &model.Message{message}}
}

func (c *CommandResult) IsFinished() bool {
	return c.Finished
}

func (c *CommandResult) GetMessage() *model.Message {
	return c.Message
}
