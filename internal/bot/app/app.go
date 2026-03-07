package app

import (
	"context"
	"fmt"
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
	scrapperService := application.NewScrapperService(scrapper)

	dispatcher := BuildDispatcher(scrapperService)
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
			err := a.processTgUpdate(ctx, update)
			if err != nil {
				slog.Error("Error processing telegram update", "error", err)
			}
		case linkUpdate, ok := <-linkUpdates:
			if !ok {
				linkUpdates = nil
				slog.Info("Link Update channel closed")
				continue
			}
			err := a.processLinkUpdate(ctx, linkUpdate)
			if err != nil {
				slog.Error("Error processing link update", "error", err)
			}
		}
		if tgUpdates == nil && linkUpdates == nil {
			slog.Info("Telegram Updates and link updates channels closed")
			break
		}
	}
}

func (a *App) processTgUpdate(ctx *context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return fmt.Errorf("telegram update message is nil")
	}

	serverResponse, err := a.dispatcher.Dispatch(ctx, update.Message.Text, update.Message.Chat.ID)
	if err != nil {
		return fmt.Errorf("Dispatching error", "err", err)
	}

	err = a.adapter.SendMessage(update.Message.Chat.ID, serverResponse)
	if err != nil {
		return fmt.Errorf("TG API sending message error", "err", err)
	}
	return nil
}

func (a *App) processLinkUpdate(ctx *context.Context, update domain.LinkUpdate) error {
	newMsg := application.NewMessage("Update for link " + update.Url + " new event: " + update.Description)
	for _, id := range update.TgChatIds {
		err := a.adapter.SendMessage(id, newMsg)
		if err != nil {
			return fmt.Errorf("error sending message to telegram user %d. error: %v", id, err)
		}
	}
	return nil
}

func BuildDispatcher(scrapperService *application.ScrapperService) *application.Dispatcher {
	help := &commands.HelpCommand{}
	start := &commands.StartCommand{ScrapperService: scrapperService}
	fallback := &commands.FallBackCommand{}
	track := &commands.TrackCommand{ScrapperService: scrapperService}
	untrack := &commands.UntrackCommand{ScrapperService: scrapperService}
	list := &commands.ListCommand{ScrapperService: scrapperService}
	cancel := &commands.CancelCommand{}

	cmds := []application.Command{help, start, fallback, track, untrack, list, cancel}

	d := application.NewDispatcher()
	for _, cmd := range cmds {
		d.Register(cmd)
	}

	return d
}
