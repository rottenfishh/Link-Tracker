package commands

import "context"

type UntrackCommand struct {
}

func (cmd *UntrackCommand) Name() string {
	return "/untrack"
}

func (cmd *UntrackCommand) Description() string {
	return "Command to stop following events from a given link"
}
func (cmd *UntrackCommand) Execute(ctx *context.Context, args []string) (*Message, error) {
	// Scrapper.unsubscribe
	return NewMessage("Successfully stopped tracking a given link"), nil
}
