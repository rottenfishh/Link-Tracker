package bot

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgAdapter struct {
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
}

func NewTgAdapter() *TgAdapter {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_API_TOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)

	updateConfig.Timeout = 30

	return &TgAdapter{bot: bot, updateConfig: updateConfig}
}

// TODO: different message type? (i.e. photo)
func (a *TgAdapter) SendMessage(update tgbotapi.Update, message *Message) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message.Text)
	_, err := a.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
