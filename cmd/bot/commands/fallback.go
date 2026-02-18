package commands

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot"
)

type FallBackCommand struct {
}

func (cmd *FallBackCommand) Name() string {
	return "/fallback"
}

func (cmd *FallBackCommand) Execute(ctx *context.Context, args []string) (*bot.Message, error) {
	return bot.NewMessage("Неизвестная команда. Введите \\help для просмотра доступных команд"), nil
}
