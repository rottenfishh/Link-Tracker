package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app.InitLogging()

	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("Error loading config ", "error: ", err)
		return
	}

	slog.Info("bot config loaded", "config", cfg)
	appBot, err := app.BuildApp(ctx, cfg)
	if err != nil {
		slog.Error("Error building app bot", "error", err)
		return
	}
	appBot.RunService(ctx)

	slog.Info("Shutting down...")
	<-ctx.Done()
	stop()
}
