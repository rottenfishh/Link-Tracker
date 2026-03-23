package main

import (
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/app"
)

func main() {

	config, err := app.LoadConfig()
	if err != nil {
		slog.Error("error while loading config for scheduler", "error", err)
	}

	scrapper, err := app.NewApp(config)
	if err != nil {
		slog.Error("Error creating scrapper app ", "error: ", err)
	}

	err = scrapper.Run()
	if err != nil {
		slog.Error("Error starting scrapper app ", "error: ", err)
	}
}
