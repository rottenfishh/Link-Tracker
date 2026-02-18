package commands

import (
	"context"
)

type FallBackCommand struct {
}

func (cmd *FallBackCommand) Name() string {
	return "/fallback"
}

func (cmd *FallBackCommand) Description() string {
	return "Fallback for unknown command"
}
func (cmd *FallBackCommand) Execute(ctx *context.Context, args []string) (*Message, error) {
	return NewMessage("Неизвестная команда. Введите \\help для просмотра доступных команд"), nil
}
