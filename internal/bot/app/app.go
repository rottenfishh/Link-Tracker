package app

import (
	"context"
	"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application/commands"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure/in"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type App struct {
	dispatcher *application.Dispatcher
	adapter    *infrastructure.TgClient
	server     *in.Server
	config     AppConfig
}

func NewApp() *App {
	InitLogging()

	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("Error loading config ", "error: ", err)
		return nil
	}

	scrapper := infrastructure.NewScrapperClient(cfg.ScrapperUrl)
	dispatcher := BuildDispatcher(scrapper)
	adapter := infrastructure.NewTgAdapter(cfg.Telegram.Token, cfg.Telegram.Debug)

	slog.Info("Finished setting up service")

	router := in.NewServer(":" + strconv.Itoa(cfg.Port))

	return &App{dispatcher, adapter, router, *cfg}
}

// TODO: FAN-IN pattern for accepting events from both tg and http channels
// TODO: idk how to do this
func (a *App) RunService(ctx *context.Context) {
	slog.Info("Starting service. Accepting user messages")

	go func() {
		err := a.server.Run()
		if err != nil {
			slog.Error("Error starting http server", "error", err)
		}
	}()

	linkUpdates := a.server.GetUpdates()
	tgUpdates := a.adapter.GetUpdates()
	for {
		select {
		case update, ok := <-tgUpdates:
			if !ok {
				tgUpdates = nil
				slog.Info("Telegram Update channel closed")
				continue
			}
			a.processTgUpdate(ctx, update)
		case linkUpdate, ok := <-linkUpdates:
			if !ok {
				linkUpdates = nil
				slog.Info("Link Update channel closed")
				continue
			}
			a.processLinkUpdate(ctx, linkUpdate)
		}
		if tgUpdates == nil && linkUpdates == nil {
			break
		}
	}
	for update := range tgUpdates {
		if update.Message == nil {
			continue
		}

		serverResponse, err := a.dispatcher.Dispatch(ctx, update.Message.Text, update.Message.Chat.ID)
		if err != nil {

			slog.Error("Dispatching error", "err", err)
		}

		err = a.adapter.SendMessage(update, serverResponse)
		if err != nil {
			slog.Error("TG API sending message error", "err", err)
		}
	}
}

func (a *App) processTgUpdate(ctx *context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	serverResponse, err := a.dispatcher.Dispatch(ctx, update.Message.Text, update.Message.Chat.ID)
	if err != nil {
		slog.Error("Dispatching error", "err", err)
		return
	}

	err = a.adapter.SendMessage(update, serverResponse)
	if err != nil {
		slog.Error("TG API sending message error", "err", err)
		return
	}
}

func (a *App) processLinkUpdate(ctx *context.Context, update domain.LinkUpdate) {

}
func BuildDispatcher(scrapperCl infrastructure.ScrapperClient) *application.Dispatcher {
	help := &commands.HelpCommand{}
	start := &commands.StartCommand{}
	fallback := &commands.FallBackCommand{}
	track := &commands.TrackCommand{ScrapperClient: scrapperCl}
	untrack := &commands.UntrackCommand{}
	list := &commands.ListCommand{}
	cancel := &commands.CancelCommand{}

	cmds := []application.Command{help, start, fallback, track, untrack, list, cancel}

	d := application.NewDispatcher()
	for _, cmd := range cmds {
		d.Register(cmd)
	}

	return d
}
