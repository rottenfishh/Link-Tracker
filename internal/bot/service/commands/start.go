package commands

import (
	"context"
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/state"
)

type StartCommand struct {
	ScrapperClient ScrapperClient
}

func NewStartCommand(scrapperClient ScrapperClient) *StartCommand {
	return &StartCommand{scrapperClient}
}

func (cmd *StartCommand) Name() string {
	return "/start"
}

func (cmd *StartCommand) Description() string {
	return "Start working with bot"
}

func (cmd *StartCommand) Execute(ctx context.Context, state *state.State) (*CommandResult, error) {
	err := cmd.ScrapperClient.RegisterChat(ctx, state.ChatID)
	if err != nil {
		return nil, fmt.Errorf("registering chat in scrapper: %w", err)
	}
	return NewCommandResult(true, "Добро пожаловать, путник. Введите /help для просмотра доступных команд."), nil
}
