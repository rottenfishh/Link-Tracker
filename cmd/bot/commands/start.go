package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot"
)

type StartCommand struct{}

func (cmd *StartCommand) Name() string {
	return "/start"
}

func (cmd *StartCommand) Execute(ctx *context.Context, args []string) (*bot.Message, error) {
	return bot.NewMessage("Добро пожаловать, путник"), nil
}
