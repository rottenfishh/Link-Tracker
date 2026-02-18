package bot

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()

	adapter := NewTgAdapter()
	updates := adapter.bot.GetUpdatesChan(adapter.updateConfig)

	dispatcher := NewDispatcher()

	for update := range updates {
		if update.Message == nil {
			continue
		}
		serverResponse, err := dispatcher.Dispatch(&ctx, update.Message.Text)
		if err != nil {
			fmt.Println(err)
		}

		err = adapter.SendMessage(update, serverResponse)
		if err != nil {
			fmt.Println(err)
		}
	}
}
