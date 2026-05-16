package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := app.LoadConfig()
	if err != nil {
		slog.Error("error while loading config for scheduler", "error", err)
	}
	slog.Info("load config", "config", config)

	scrapper, err := app.BuildApp(ctx, config)
	if err != nil {
		slog.Error("Error creating scrapper app ", "error: ", err)
	}

	err = scrapper.Run(ctx)
	if err != nil {
		slog.Error("Error starting scrapper app ", "error: ", err)
	}

	slog.Info("Shutting down...")
	<-ctx.Done()
	stop()
}
