package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
)

type CancelCommand struct {
}

func (c *CancelCommand) Name() string {
	return "/cancel"
}

func (c *CancelCommand) Description() string {
	return "Command to cancel current dialogue"
}

func (c *CancelCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	return application.NewCommandResult(true, "Завершение текущего диалога..."), nil
}
