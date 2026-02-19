package app

import (
	"context"
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application/commands"
	infrastructure2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure"
)

type App struct {
	dispatcher *application.Dispatcher
	adapter    *infrastructure2.TgAdapter
}

func NewApp() *App {
	err := LoadEnv()

	if err != nil {
		fmt.Printf("Load env err: %v", err)
	}
	d := buildDispatcher()
	a := infrastructure2.NewTgAdapter()
	return &App{d, a}
}

func buildDispatcher() *application.Dispatcher {
	help := &commands.HelpCommand{}
	start := &commands.StartCommand{}
	fallback := &commands.FallBackCommand{}
	cmds := []commands.Command{help, start, fallback}

	d := application.NewDispatcher()
	for _, cmd := range cmds {
		d.Register(cmd)
	}

	return d
}

func (a *App) RunService(ctx *context.Context) {
	updates := a.adapter.Bot.GetUpdatesChan(a.adapter.UpdateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		serverResponse, err := a.dispatcher.Dispatch(ctx, update.Message.Text)
		if err != nil {
			fmt.Println(err)
		}

		err = a.adapter.SendMessage(update, serverResponse)
		if err != nil {
			fmt.Println(err)
		}
	}
}
