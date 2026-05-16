package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/state"
)

type HelpCommand struct {
}

func (cmd *HelpCommand) Name() string {
	return "/help"
}

func (cmd *HelpCommand) Description() string {
	return "Show help: bot's available commands"
}

func (cmd *HelpCommand) Execute(ctx context.Context, state *state.State) (*CommandResult, error) {
	_ = ctx
	_ = state
	return NewCommandResult(true, "Для начала работы, введите сначала /start\n"+
		"Доступные в боте команды:\n"+
		"/start - начать работу\n"+
		"/help - справка по боту\n"+
		"/track - отслеживать ссылку\n"+
		"/untrack <link> - перестать отслеживать ссылку\n"+
		"/list - посмотреть свои отслеживаемые ссылки\n"+
		"/cancel - отменить текущую команду\n"), nil

}
