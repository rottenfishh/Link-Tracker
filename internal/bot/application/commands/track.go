package commands

import (
	"context"
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
)

type TrackCommand struct {
}

func (cmd *TrackCommand) Name() string {
	return "/track"
}

func (cmd *TrackCommand) Description() string {
	return "Command to start tracking events from a given link"
}

// TODO: check for no args
// get link. after that adapter sets flag for start of state machine(int ctx, for example)
// . if flag is set, accept args as tags. /cancel is processed in adapter, and sent to this cmmand to. if its sent, we save stuff
func (cmd *TrackCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	// call Scrapper.RegisterLink
	// check ifs not already tracked
	//После ввода /track бот получает ссылку на отслеживаемый ресурс.
	//    После получения ссылки бот запрашивает теги (необязательно). Теги -- это слова или короткие фразы, разделённые запятыми, по которым пользователь может классифицировать ресурс.
	//    Примеры: работа, баг, документация, рефакторинг.
	//    Пользователь может ввести /cancel, чтобы прервать процесс.
	//    Если пользователь отправляет другую команду — процесс отслеживания отменяется.
	if state.Step == 1 {
		if len(state.UserArgs) == 0 {
			return application.NewCommandResult(false, "Пожалуйста, введите команду в виде /track ваша-ссылка"), nil
		}
		state.Data["link"] = state.UserArgs[0]
		state.Step = 2
		// TODO: check if its already tracked in scrapper
		return application.NewCommandResult(false, "Пожалуйста, введи теги для вашей ссылки: "+state.UserArgs[0]), nil
	}
	if state.Step == 2 {
		link := state.Data["link"].(string)
		if len(state.UserArgs) == 0 {
			// TODO: Scrapper.RegisterLink(link)
			return application.NewCommandResult(true, "Успешно начали отслеживание ссылки "+link), nil
		}
		tags := state.UserArgs
		// TODO: Scapprer.RegisterLink(link, tags). all tags
		return application.NewCommandResult(true, "Успешно начали отслеживание ссылки "+link+" с тегами"+tags[0]), nil
	}

	return application.NewCommandResult(true, "Неизвестная команда"), fmt.Errorf("unknown command %v", state.UserArgs)
}
