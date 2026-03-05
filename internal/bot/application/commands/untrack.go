package commands

import (
	"context"
	"errors"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type UntrackCommand struct {
	ScrapperService *application.ScrapperService
}

func (cmd *UntrackCommand) Name() string {
	return "/untrack"
}

func (cmd *UntrackCommand) Description() string {
	return "Command to stop following events from a given link"
}
func (cmd *UntrackCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	// Scrapper.unsubscribe
	if len(state.UserArgs) == 0 {
		return application.NewCommandResult(true, "You need to provide a link to untrack"), nil
	}

	err := cmd.ScrapperService.DeleteLink(state.ChatId, state.UserArgs[0])
	if err != nil {
		var msg string
		slog.Error("Unregistering link error ", "chatId", state.ChatId, "link", state.UserArgs[0], "error", err)
		switch {
		case errors.Is(err, domain.ErrNotFound):
			msg = "Ссылка или чат не найдены"
		case errors.Is(err, domain.ErrInvalidRequest):
			msg = "Некорректные параметры запроса"
		}
		return application.NewCommandResult(true, msg), nil
	}
	return application.NewCommandResult(true, "Successfully stopped tracking a given link"), nil
}
