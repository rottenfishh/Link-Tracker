package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

// т.н. dependency injection
type TrackCommand struct {
	ScrapperService *application.ScrapperService
}

func (cmd *TrackCommand) Name() string {
	return "/track"
}

func (cmd *TrackCommand) Description() string {
	return "Command to start tracking events from a given link"
}

// TODO: inline buttons for skipping tags
// get link. after that adapter sets flag for start of state machine(int ctx, for example)
// . if flag is set, accept args as tags. /cancel is processed in adapter, and sent to this cmmand to. if its sent, we save stuff
func (cmd *TrackCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	if state.Step == 1 {
		if len(state.UserArgs) == 0 {
			return application.NewCommandResult(false, "Пожалуйста, введите команду в виде /track ваша-ссылка"), nil
		}
		state.Data["link"] = state.UserArgs[0]
		state.Step = 2
		// TODO: check if its already tracked in scrapper
		return application.NewCommandResult(false, "Пожалуйста, введи теги для вашей ссылки. "+
			"Введите \"-\" для сохранения без тегов"), nil
	}
	if state.Step == 2 {
		link := state.Data["link"].(string)

		if state.UserArgs[0] == "-" {
			state.UserArgs = nil
		}

		err := cmd.ScrapperService.AddLink(state.ChatId, link, state.UserArgs)
		if err != nil {
			slog.Error("Registering link error ", "link", link, "error", err)
			var msg string
			switch {
			case errors.Is(err, domain.ErrLinkAlreadyTracked):
				msg = "Ссылка уже отслеживается"
			case errors.Is(err, domain.ErrNotFound):
				msg = "Чат не существует"
			case errors.Is(err, domain.ErrInvalidRequest):
				msg = "Неправильное параметры команды. Пожалуйста, введите ссылку"
			default:
				msg = "Не удалось начать отслеживать ссылку"
			}
			return application.NewCommandResult(true, msg), nil
		}

		return application.NewCommandResult(true, "Успешно начали отслеживание ссылки "+link), nil
	}

	return application.NewCommandResult(true, "Неизвестная команда"), fmt.Errorf("unknown command %v", state.UserArgs)
}
