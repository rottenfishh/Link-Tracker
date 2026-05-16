package app

import (
	"context"
	"fmt"
	"log/slog"

	inkafka "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/adapter/in/kafka"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/adapter/out/ai"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/adapter/out/kafka"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/agent/service"
)

type App struct {
	cfg            Config
	updateConsumer *inkafka.Consumer
}

func BuildApp(cfg Config) (*App, error) {
	producer, err := kafka.NewProducer(cfg.ProducerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating kafka producer: %w", err)
	}
	summarizer := ai.NewHuggingSummarizer(cfg.HuggingToken)
	filter := service.NewFilter(cfg.FilterConfig, summarizer, producer)

	consumer, err := inkafka.NewConsumer(cfg.ConsumerConfig, filter)
	if err != nil {
		slog.Error("failed to build kafka consumer", "error", err)
		return nil, fmt.Errorf("failed to build kafka consumer %w", err)
	}

	return &App{cfg: cfg, updateConsumer: consumer}, nil
}

func (a *App) Run(ctx context.Context) error {
	err := a.updateConsumer.Run(ctx)

	if err != nil {
		return fmt.Errorf("starting consumer failed: %w", err)
	}
	return nil
}
