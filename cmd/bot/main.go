package main

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/app"
)

func main() {
	ctx := context.Background()

	appBot := app.NewApp()

	appBot.RunService(&ctx)
}
