package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/state"
)

type FallBackCommand struct {
}

func (cmd *FallBackCommand) Name() string {
	return "/fallback"
}

func (cmd *FallBackCommand) Description() string {
	return "Fallback for unknown command"
}

func (cmd *FallBackCommand) Execute(ctx context.Context, state *state.State) (*CommandResult, error) {
	_ = ctx
	_ = state
	return NewCommandResult(true, "Неизвестная команда. Введите /help для просмотра доступных команд."), nil
}
