package main

import (
	"context"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, err := app.LoadConfig()
	if err != nil {
		slog.Error("Error loading config: ", err)
		return
	}
	slog.Info("Load config success", "config", config)
	agentApp, err := app.BuildApp(*config)
	if err != nil {
		slog.Error("Error building app: ", err)
		return
	}

	slog.Info("AI service starting")
	if err = agentApp.Run(ctx); err != nil {
		slog.Error("Error running app: ", err)
		return
	}
}
