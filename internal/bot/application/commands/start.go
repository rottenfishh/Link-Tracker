package commands

import (
	"context"
)

type StartCommand struct{}

func (cmd *StartCommand) Name() string {
	return "/start"
}

func (cmd *StartCommand) Description() string {
	return "Start working with bot"
}

func (cmd *StartCommand) Execute(ctx *context.Context, args []string) (*Message, error) {
	return NewMessage("Добро пожаловать, путник. Введите /help для просмотра доступных команд."), nil
}
