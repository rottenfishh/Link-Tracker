package service

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/commands"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/state"
)

type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, state *state.State) (*commands.CommandResult, error)
}
