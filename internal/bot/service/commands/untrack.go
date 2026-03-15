package commands

import (
	"context"
	"errors"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type UntrackCommand struct {
	ScrapperService *service.ScrapperService
}

func (cmd *UntrackCommand) Name() string {
	return "/untrack"
}

func (cmd *UntrackCommand) Description() string {
	return "Command to stop following events from a given link"
}

func (cmd *UntrackCommand) Execute(ctx context.Context, state *service.State) (*service.CommandResult, error) {
	if len(state.UserArgs) == 0 {
		return service.NewCommandResult(true, "You need to provide a link to untrack"), nil
	}

	err := cmd.ScrapperService.DeleteLink(ctx, state.ChatId, state.UserArgs[0])
	if err != nil {
		var msg string
		slog.Error("Unregistering link error ", "chatId", state.ChatId, "link", state.UserArgs[0], "error", err)
		switch {
		case errors.Is(err, model.ErrNotFound):
			msg = "Ссылка или чат не найдены"
		case errors.Is(err, model.ErrInvalidRequest):
			msg = "Некорректные параметры запроса"
		}
		return service.NewCommandResult(true, msg), nil
	}
	return service.NewCommandResult(true, "Successfully stopped tracking a given link"), nil
}
