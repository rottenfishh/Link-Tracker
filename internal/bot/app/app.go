package app

import (
	"context"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application/commands"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure"
)

type App struct {
	dispatcher *application.Dispatcher
	adapter    *infrastructure.TgAdapter
}

func NewApp() *App {
	InitLogging()

	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("Error loading config ", "error: ", err)
		return nil
	}
	d := BuildDispatcher()
	a := infrastructure.NewTgAdapter(cfg.Telegram.Token, cfg.Telegram.Debug)

	slog.Info("Finished setting up service")
	return &App{d, a}
}

func BuildDispatcher() *application.Dispatcher {
	help := &commands.HelpCommand{}
	start := &commands.StartCommand{}
	fallback := &commands.FallBackCommand{}
	track := &commands.TrackCommand{}
	untrack := &commands.UntrackCommand{}
	list := &commands.ListCommand{}

	cmds := []commands.Command{help, start, fallback, track, untrack, list}

	d := application.NewDispatcher()
	for _, cmd := range cmds {
		d.Register(cmd)
	}

	return d
}

func (a *App) RunService(ctx *context.Context) {
	slog.Info("Starting service. Accepting user messages")

	updates := a.adapter.Bot.GetUpdatesChan(a.adapter.UpdateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		serverResponse, err := a.dispatcher.Dispatch(ctx, update.Message.Text)
		if err != nil {

			slog.Error("Dispatching error", "err", err)
		}

		err = a.adapter.SendMessage(update, serverResponse)
		if err != nil {
			slog.Error("TG API sending message error", "err", err)
		}
	}
}
