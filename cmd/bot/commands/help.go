package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot"
)

type HelpCommand struct {
}

func (cmd *HelpCommand) Name() string {
	return "/help"
}
func (cmd *HelpCommand) Execute(ctx *context.Context, args []string) (*bot.Message, error) {
	return bot.NewMessage("Доступные в боте команды:\n" +
		"\\start - начать работу\n" +
		"\\help - справка по боту"), nil
}
