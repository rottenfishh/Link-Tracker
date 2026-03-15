package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
)

type StartCommand struct {
	ScrapperService *service.ScrapperService
}

func (cmd *StartCommand) Name() string {
	return "/start"
}

func (cmd *StartCommand) Description() string {
	return "Start working with bot"
}

func (cmd *StartCommand) Execute(ctx context.Context, state *service.State) (*service.CommandResult, error) {
	err := cmd.ScrapperService.RegisterChat(ctx, state.ChatId)
	if err != nil {
		return nil, err
	}
	return service.NewCommandResult(true, "Добро пожаловать, путник. Введите /help для просмотра доступных команд."), nil
}
