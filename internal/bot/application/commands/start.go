package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
)

type StartCommand struct {
	ScrapperService *application.ScrapperService
}

func (cmd *StartCommand) Name() string {
	return "/start"
}

func (cmd *StartCommand) Description() string {
	return "Start working with bot"
}

func (cmd *StartCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	err := cmd.ScrapperService.RegisterChat(state.ChatId)
	if err != nil {
		return nil, err
	}
	return application.NewCommandResult(true, "Добро пожаловать, путник. Введите /help для просмотра доступных команд."), nil
}
