package main

import (
	"context"
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/app"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/infrastructure"
)

func main() {
	ctx := context.Background()

	err := app.LoadEnv()
	if err != nil {
		fmt.Printf("Load env err: %v", err)
	}
	adapter := infrastructure.NewTgAdapter()
	updates := adapter.Bot.GetUpdatesChan(adapter.UpdateConfig)

	dispatcher := app.BuildDispatcher()

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
