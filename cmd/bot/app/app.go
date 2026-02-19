package app

import (
	"context"
	"fmt"

	commands2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/application/commands"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/infrastructure"
)

type App struct {
	dispatcher *infrastructure.Dispatcher
	adapter    *infrastructure.TgAdapter
}

func NewApp() *App {
	err := LoadEnv()

	if err != nil {
		fmt.Printf("Load env err: %v", err)
	}
	d := buildDispatcher()
	a := infrastructure.NewTgAdapter()
	return &App{d, a}
}

func buildDispatcher() *infrastructure.Dispatcher {
	help := &commands2.HelpCommand{}
	start := &commands2.StartCommand{}
	fallback := &commands2.FallBackCommand{}
	cmds := []commands2.Command{help, start, fallback}

	d := infrastructure.NewDispatcher()
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
