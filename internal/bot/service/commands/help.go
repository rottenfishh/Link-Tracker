package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
)

type HelpCommand struct {
}

func (cmd *HelpCommand) Name() string {
	return "/help"
}

func (cmd *HelpCommand) Description() string {
	return "Show help: bot's available commands"
}

func (cmd *HelpCommand) Execute(ctx context.Context, state *service.State) (*service.CommandResult, error) {
	return service.NewCommandResult(true, "Доступные в боте команды:\n"+
		"/start - начать работу\n"+
		"/help - справка по боту"), nil
}
