package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
)

type HelpCommand struct {
}

func (cmd *HelpCommand) Name() string {
	return "/help"
}
func (cmd *HelpCommand) Description() string {
	return "Show help: bot's available commands"
}

func (cmd *HelpCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	return application.NewCommandResult(true, "Доступные в боте команды:\n"+
		"/start - начать работу\n"+
		"/help - справка по боту"), nil
}
