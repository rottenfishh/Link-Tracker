package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
)

type ListCommand struct {
}

func (cmd *ListCommand) Name() string {
	return "/list"
}

func (cmd *ListCommand) Description() string {
	return "List all tracked events"
}

func (cmd *ListCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	// Scrapper.getSubscribed(userId, <Optional> tags)
	list := make([]string, 0) // -
	if len(list) == 0 {
		return application.NewCommandResult(true, "Вы пока не отслеживаете ни одной ссылки"), nil
	}
	return application.NewCommandResult(true, "arr"), nil
}
