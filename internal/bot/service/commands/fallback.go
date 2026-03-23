package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
)

type FallBackCommand struct {
}

func (cmd *FallBackCommand) Name() string {
	return "/fallback"
}

func (cmd *FallBackCommand) Description() string {
	return "Fallback for unknown command"
}

func (cmd *FallBackCommand) Execute(ctx context.Context, state *service.State) (*service.CommandResult, error) {
	return service.NewCommandResult(true, "Неизвестная команда. Введите /help для просмотра доступных команд."), nil
}
