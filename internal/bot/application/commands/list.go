package commands

import "context"

type ListCommand struct {
}

func (cmd *ListCommand) Name() string {
	return "/list"
}

func (cmd *ListCommand) Description() string {
	return "List all tracked events"
}

func (cmd *ListCommand) Execute(ctx *context.Context, args []string) (*Message, error) {
	// Scrapper.getSubscribed(userId, <Optional> tags)
	list := make([]string, 0) // -
	if len(list) == 0 {
		return NewMessage("Вы пока не отслеживаете ни одной ссылки"), nil
	}
	return NewMessage("arr"), nil
}
