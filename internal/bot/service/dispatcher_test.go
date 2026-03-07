package service

import (
	"context"
	"testing"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/commands"
)

func TestStartCommand(t *testing.T) {
	d := NewDispatcher()
	start := &commands.StartCommand{}
	fallback := &commands.FallBackCommand{}
	d.Register(start)
	d.Register(fallback)

	ctx := context.Background()
	msg, err := d.Dispatch(ctx, "/start", 0)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Добро пожаловать, путник. Введите /help для просмотра доступных команд."
	if msg.Text != expected {
		t.Fatalf("ожидалось %s, получили %s", expected, msg.Text)
	}
}

func TestHelpCommand(t *testing.T) {
	d := NewDispatcher()
	help := &commands.HelpCommand{}
	fallback := &commands.FallBackCommand{}
	d.Register(help)
	d.Register(fallback)

	ctx := context.Background()
	msg, err := d.Dispatch(ctx, "/help", 0)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Доступные в боте команды:\n" +
		"/start - начать работу\n" +
		"/help - справка по боту"

	if msg.Text != expected {
		t.Fatalf("ожидалось %s, получили %s", expected, msg.Text)
	}
}

func TestFallbackCommand(t *testing.T) {
	d := NewDispatcher()
	start := &commands.StartCommand{}
	fallback := &commands.FallBackCommand{}
	d.Register(start)
	d.Register(fallback)

	ctx := context.Background()
	msg, err := d.Dispatch(ctx, "bebebe", 0)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Неизвестная команда. Введите /help для просмотра доступных команд."
	if msg.Text != expected {
		t.Fatalf("ожидалось %s, получили %s", expected, msg.Text)
	}
}
