package commands

import (
	"context"
)

type HelpCommand struct {
}

func (cmd *HelpCommand) Name() string {
	return "/help"
}
func (cmd *HelpCommand) Description() string {
	return "Show help: bot's available commands"
}

func (cmd *HelpCommand) Execute(ctx *context.Context, args []string) (*Message, error) {
	return NewMessage("Доступные в боте команды:\n" +
		"\\start - начать работу\n" +
		"\\help - справка по боту"), nil
}
