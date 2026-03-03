package infrastructure

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/application"
)

type TgClient struct {
	Bot          *tgbotapi.BotAPI
	UpdateConfig tgbotapi.UpdateConfig
}

func NewTgAdapter(key string, debug bool) *TgClient {
	slog.Info("Initializing TgClient")

	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		slog.Error("Error setting up API bot", "error", err)
	}

	bot.Debug = debug
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	SetUpTgCommands(bot)

	return &TgClient{Bot: bot, UpdateConfig: updateConfig}
}

// TODO: use command structs to extract this info?
func SetUpTgCommands(bot *tgbotapi.BotAPI) {
	cmds := []tgbotapi.BotCommand{{
		Command: "start", Description: "start the bot to use our cool features"},
		{Command: "help", Description: "show bot's available commands"},
		{Command: "track", Description: "track given link"},
		{Command: "untrack", Description: "untrack given link"},
		{Command: "list", Description: "get list of tracked links"},
		{Command: "cancel", Description: "cancel current command execution"},
	}
	cfg := tgbotapi.NewSetMyCommands(cmds...)
	_, err := bot.Request(cfg)
	if err != nil {
		slog.Error("Error while setting up tg commands", "error", err)
	}
}

// TODO: different message type? (i.e. photo)
func (a *TgClient) SendMessage(update tgbotapi.Update, message *application.Message) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message.Text)
	_, err := a.Bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (a *TgClient) GetUpdates() tgbotapi.UpdatesChannel {
	return a.Bot.GetUpdatesChan(a.UpdateConfig)
}
