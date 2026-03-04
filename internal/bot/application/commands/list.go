package commands

import (
	"context"
	"log/slog"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure"
)

type ListCommand struct {
	ScrapperCLient infrastructure.ScrapperClient
}

func (cmd *ListCommand) Name() string {
	return "/list"
}

func (cmd *ListCommand) Description() string {
	return "List all tracked events"
}

func (cmd *ListCommand) Execute(ctx *context.Context, state *application.State) (*application.CommandResult, error) {
	list, err := cmd.ScrapperCLient.GetLinks(state.ChatId)
	if err != nil {
		slog.Error("Error getting links for chat", "id", state.ChatId)
		return application.NewCommandResult(true, "Не удалось получить ссылки"), nil
	}
	var res strings.Builder
	for _, link := range list {
		res.WriteString(link.Link + "\n")
	}
	if len(list) == 0 {
		return application.NewCommandResult(true, "Вы пока не отслеживаете ни одной ссылки"), nil
	}
	return application.NewCommandResult(true, res.String()), nil
}
