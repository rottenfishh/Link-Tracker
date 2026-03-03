package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
)

type UntrackCommand struct {
}

func (cmd *UntrackCommand) Name() string {
	return "/untrack"
}

func (cmd *UntrackCommand) Description() string {
	return "Command to stop following events from a given link"
}
func (cmd *UntrackCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	// Scrapper.unsubscribe
	return application.NewCommandResult(true, "Successfully stopped tracking a given link"), nil
}
