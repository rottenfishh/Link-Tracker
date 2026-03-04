package commands

import (
	"context"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure"
)

type UntrackCommand struct {
	ScrapperClient infrastructure.ScrapperClient
}

func (cmd *UntrackCommand) Name() string {
	return "/untrack"
}

func (cmd *UntrackCommand) Description() string {
	return "Command to stop following events from a given link"
}
func (cmd *UntrackCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	// Scrapper.unsubscribe
	err := cmd.ScrapperClient.DeleteLink(state.ChatId, state.UserArgs[0])
	if err != nil {
		slog.Error("Unregistering link error ", "chatId", state.ChatId, "link", state.UserArgs[0], "error", err)
		return nil, err
	}
	return application.NewCommandResult(true, "Successfully stopped tracking a given link"), nil
}
