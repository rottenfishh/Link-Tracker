package commands

import (
	"context"
	"log/slog"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
)

type ListCommand struct {
	ScrapperService *service.ScrapperService
}

func (cmd *ListCommand) Name() string {
	return "/list"
}

func (cmd *ListCommand) Description() string {
	return "List all tracked events"
}

func (cmd *ListCommand) Execute(ctx context.Context, state *service.State) (*service.CommandResult, error) {
	list, err := cmd.ScrapperService.GetLinks(ctx, state.ChatId)
	if err != nil {
		slog.Error("Error getting links for chat", "id", state.ChatId)
		return service.NewCommandResult(true, "Не удалось получить ссылки"), nil
	}
	var res strings.Builder
	for _, link := range list {
		res.WriteString(link.Link + "\n")
	}
	if len(list) == 0 {
		return service.NewCommandResult(true, "Вы пока не отслеживаете ни одной ссылки"), nil
	}
	return service.NewCommandResult(true, res.String()), nil
}
