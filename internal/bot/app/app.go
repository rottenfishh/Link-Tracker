package app

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application/commands"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/infrastructure/in"
)

type App struct {
	dispatcher *application.Dispatcher
	adapter    *infrastructure.TgClient
	server     *gin.Engine
	config     AppConfig
}

func NewApp() *App {
	InitLogging()

	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("Error loading config ", "error: ", err)
		return nil
	}
	dispatcher := BuildDispatcher()
	adapter := infrastructure.NewTgAdapter(cfg.Telegram.Token, cfg.Telegram.Debug)

	slog.Info("Finished setting up service")

	router := BuildServer()

	return &App{dispatcher, adapter, router, *cfg}
}

// TODO: FAN-IN pattern for accepting events from both tg and http channels
// TODO: idk how to do this
func (a *App) RunService(ctx *context.Context) {
	slog.Info("Starting service. Accepting user messages")

	go func() {
		err := a.server.Run(":" + strconv.Itoa(a.config.Port))
		if err != nil {
			slog.Error("Error starting http server", "error", err)
		}
	}()

	updates := a.adapter.GetUpdates()

	for update := range updates {
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

func BuildDispatcher() *application.Dispatcher {
	help := &commands.HelpCommand{}
	start := &commands.StartCommand{}
	fallback := &commands.FallBackCommand{}
	track := &commands.TrackCommand{}
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

func BuildServer() *gin.Engine {
	router := gin.Default()
	handler := in.HttpHandler{}
	registerRoutes(router, handler)
	return router
}

func registerRoutes(router *gin.Engine, handler in.HttpHandler) {
	router.POST("/updates", handler.UpdateFromLink)
}
