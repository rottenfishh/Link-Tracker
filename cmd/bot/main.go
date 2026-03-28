package main

import (
	"context"
	"log/slog"
	"os"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
)

func main() {
	ctx := context.Background()

	app.InitLogging()

	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("Error loading config ", "error: ", err)
		os.Exit(1)
	}

	appBot := app.NewApp(cfg)
	appBot.RunService(ctx)
}
