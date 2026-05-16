//nolint:mnd // this command is a tiny state machine with explicit step numbers
package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/state"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

// TrackCommand handles the multi-step flow for subscribing to a link.
type TrackCommand struct {
	ScrapperClient ScrapperClient
}

func NewTrackCommand(scrapperClient ScrapperClient) *TrackCommand {
	return &TrackCommand{ScrapperClient: scrapperClient}
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
func (cmd *TrackCommand) Execute(ctx context.Context, state *state.State) (*CommandResult, error) {
	if state.Step == 1 {
		state.Step = 2
		return NewCommandResult(false, "Пожалуйста, введите ссылку"), nil
	}

	if state.Step == 2 {
		if len(state.UserArgs) == 0 {
			return NewCommandResult(false, "Пожалуйста, введите ссылку"), nil
		}
		state.Data["link"] = state.UserArgs[0]
		state.Step = 3
		// TODO: check if its already tracked in scrapper
		return NewCommandResult(false, "Пожалуйста, введи теги для вашей ссылки. "+
			"Введите \"-\" для сохранения без тегов"), nil
	}
	if state.Step == 3 {
		link, _ := state.Data["link"].(string)

		if state.UserArgs[0] == "-" {
			state.UserArgs = nil
		}

		err := cmd.ScrapperClient.RegisterLink(ctx, state.ChatID, link, state.UserArgs)
		if err != nil {
			slog.Error("Registering link error ", "link", link, "error", err)
			var msg string
			switch {
			case errors.Is(err, model.ErrLinkAlreadyTracked):
				msg = "Ссылка уже отслеживается"
			case errors.Is(err, model.ErrNotFound):
				msg = "Чат не существует"
			case errors.Is(err, model.ErrInvalidRequest):
				msg = "Неверная или неподдерживаемая ссылка. Введите ссылку с github или stackoverflow"
			default:
				msg = "Не удалось начать отслеживать ссылку"
			}
			return NewCommandResult(true, msg), nil
		}

		return NewCommandResult(true, "Успешно начали отслеживание ссылки "+link), nil
	}

	return NewCommandResult(true, "Неизвестная команда"), fmt.Errorf("unknown command %v", state.UserArgs)
}
