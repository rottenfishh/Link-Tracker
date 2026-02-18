package app

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/commands"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/infrastructure"
)

func BuildDispatcher() *infrastructure.Dispatcher {
	help := &commands.HelpCommand{}
	start := &commands.StartCommand{}
	fallback := &commands.FallBackCommand{}
	cmds := []commands.Command{help, start, fallback}

	d := infrastructure.NewDispatcher()
	for _, cmd := range cmds {
		d.Register(cmd)
	}

	return d
}
