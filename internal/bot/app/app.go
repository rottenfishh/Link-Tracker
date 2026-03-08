package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in"
	grpc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/http"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/commands"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type App struct {
	dispatcher *service.Dispatcher
	adapter    in.TgClient
	server     *http.Server
	grpcServer *grpc.BotServer
	config     *AppConfig
}

func NewApp(cfg *AppConfig) *App {
	scrapper := out.NewScrapperClient(cfg.ScrapperUrl)
	scrapperService := service.NewScrapperService(scrapper)

	dispatcher := BuildDispatcher(scrapperService)
	adapter := in.NewTgAdapter(cfg.Telegram.Token, cfg.Telegram.Debug)

	slog.Info("Finished setting up service")

	publisher := service.NewUpdatePublisher()
	grpcServer := grpc.NewBotServiceServer(publisher)
	//router := http.NewServer(":"+strconv.Itoa(cfg.Port), publisher)

	return &App{dispatcher: dispatcher, adapter: adapter, config: cfg, grpcServer: grpcServer}
}

func (a *App) RunService(ctx context.Context) {
	slog.Info("Starting service. Accepting user messages")

	go func() {
		err := a.grpcServer.RunServer("8088", strconv.Itoa(a.config.Port))
		if err != nil {
			slog.Error("Error starting http server", "error", err)
		}
	}()

	linkUpdates := a.grpcServer.GetUpdates()
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

func (a *App) processTgUpdate(ctx context.Context, update model.ChatUpdate) error {
	if update.Message == nil {
		return fmt.Errorf("telegram update message is nil")
	}

	serverResponse, err := a.dispatcher.Dispatch(ctx, update.Message.Text, update.ChatID)
	if err != nil {
		return fmt.Errorf("Dispatching error", "err", err)
	}

	err = a.adapter.SendMessage(update.ChatID, serverResponse)
	if err != nil {
		return fmt.Errorf("TG API sending message error", "err", err)
	}
	return nil
}

func (a *App) processLinkUpdate(ctx context.Context, update model.LinkUpdate) error {
	newMsg := model.NewMessage("Update for link " + update.Link + " new event: " + update.Description)
	for _, id := range update.TgChatIds {
		err := a.adapter.SendMessage(id, newMsg)
		if err != nil {
			return fmt.Errorf("error sending message to telegram user %d. error: %v", id, err)
		}
	}
	return nil
}

func BuildDispatcher(scrapperService *service.ScrapperService) *service.Dispatcher {
	help := &commands.HelpCommand{}
	start := &commands.StartCommand{ScrapperService: scrapperService}
	fallback := &commands.FallBackCommand{}
	track := &commands.TrackCommand{ScrapperService: scrapperService}
	untrack := &commands.UntrackCommand{ScrapperService: scrapperService}
	list := &commands.ListCommand{ScrapperService: scrapperService}
	cancel := &commands.CancelCommand{}

	cmds := []service.Command{help, start, fallback, track, untrack, list, cancel}

	d := service.NewDispatcher()
	for _, cmd := range cmds {
		d.Register(cmd)
	}

	return d
}
