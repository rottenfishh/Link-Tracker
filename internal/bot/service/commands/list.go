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
	var tag string
	if len(state.UserArgs) > 0 {
		tag = state.UserArgs[0]
	}

	list, err := cmd.ScrapperService.GetLinks(ctx, state.ChatId, tag)
	if err != nil {
		slog.Error("Error getting links for chat", "id", state.ChatId, "error", err)
		return service.NewCommandResult(true, "Не удалось получить ссылки"), nil
	}

	slog.Debug("Got links for chat", "id", state.ChatId, "links", list)
	var res strings.Builder
	for _, link := range list.Links {
		res.WriteString(link.Link + " : ")
		for _, linkTag := range link.Tags {
			res.WriteString(linkTag + " ")
		}
		res.WriteString("\n")
	}

	if len(list.Links) == 0 || res.String() == "\n" {
		return service.NewCommandResult(true, "Вы пока не отслеживаете ни одной ссылки"), nil
	}
	return service.NewCommandResult(true, res.String()), nil
}
