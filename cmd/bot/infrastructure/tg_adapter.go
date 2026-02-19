package infrastructure

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/application/commands"
)

type TgAdapter struct {
	Bot          *tgbotapi.BotAPI
	UpdateConfig tgbotapi.UpdateConfig
}

func NewTgAdapter() *TgAdapter {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_API_TOKEN"))
	if err != nil {
		fmt.Printf("Env: %v", err)
	}
	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	SetUpTgCommands(bot)

	return &TgAdapter{Bot: bot, UpdateConfig: updateConfig}
}

func SetUpTgCommands(bot *tgbotapi.BotAPI) {
	cmds := []tgbotapi.BotCommand{{
		Command:     "start",
		Description: "start the bot to use our cool features"},
		{
			Command:     "help",
			Description: "show bot's available commands"},
	}
	cfg := tgbotapi.NewSetMyCommands(cmds...)
	_, err := bot.Request(cfg)
	if err != nil {
		fmt.Println(err)
	}
}

// TODO: different message type? (i.e. photo)
func (a *TgAdapter) SendMessage(update tgbotapi.Update, message *commands.Message) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message.Text)
	_, err := a.Bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
