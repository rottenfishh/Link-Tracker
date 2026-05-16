//nolint:mnd // telegram long polling uses a small fixed buffer in this adapter
package in

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type TgAdapter struct {
	Bot          *tgbotapi.BotAPI
	UpdateConfig tgbotapi.UpdateConfig
}

func NewTgAdapter(key string, debug bool) *TgAdapter {
	slog.Info("Initializing TgAdapter")

	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		slog.Error("Error setting up API bot", "error", err)
	}

	bot.Debug = debug
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	SetUpTgCommands(bot)

	return &TgAdapter{Bot: bot, UpdateConfig: updateConfig}
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
func (a *TgAdapter) SendMessage(chatID int64, message *model.Message) error {
	msg := tgbotapi.NewMessage(chatID, message.Text)
	_, err := a.Bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sending telegram message: %w", err)
	}
	return nil
}

func (a *TgAdapter) GetUpdates() <-chan model.ChatUpdate {
	out := make(chan model.ChatUpdate, 100)

	go func() {
		for upd := range a.Bot.GetUpdatesChan(a.UpdateConfig) {
			if upd.Message != nil {
				out <- ToDomainChatUpdate(upd)
			}
		}
	}()

	return out
}

func ToDomainChatUpdate(update tgbotapi.Update) model.ChatUpdate {
	return model.ChatUpdate{UpdateID: int64(update.UpdateID), ChatID: update.Message.Chat.ID,
		Message: model.NewMessage(update.Message.Text)}
}
