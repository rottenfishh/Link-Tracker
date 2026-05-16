package app

import (
	"context"
	realsql "database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/out/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type App struct {
	dispatcher            *service.Dispatcher
	adapter               in.TgClient
	updatesConsumer       in.UpdatesConsumer
	updateHandlingWorkers []*out.UpdateSender
	config                *AppConfig
	repo                  service.ProcessedEventsRepository
}

// TODO: move caps to config
func BuildApp(ctx context.Context, cfg *AppConfig) (*App, error) {
	scrapper := out.NewScrapperClient(cfg.ScrapperURL, httpclient.NewReliableHTTPClient(cfg.HTTPReqConfig))

	dispatcher := BuildDispatcher(scrapper)
	adapter := in.NewTgAdapter(cfg.Telegram.Token, cfg.Telegram.Debug)

	slog.Info("Finished setting up service")

	updatesCap := 200
	reportsCap := 200
	publisher := service.NewUpdatePublisher(updatesCap, reportsCap)

	db, err := connectDB(ctx, cfg.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	processedEventsRepo := repository.NewProcessedEventsRepository(db)

	consumer, err := BuildConsumer(ctx, *cfg, publisher, processedEventsRepo)
	if err != nil {
		return nil, fmt.Errorf("error building consumer: %w", err)
	}

	workers := BuildUpdateWorkerPool(*cfg, publisher, adapter)

	return &App{
		dispatcher:            dispatcher,
		adapter:               adapter,
		config:                cfg,
		updatesConsumer:       consumer,
		updateHandlingWorkers: workers,
		repo:                  processedEventsRepo,
	}, nil
}

func connectDB(ctx context.Context, dbConfig DatabaseConfig) (*realsql.DB, error) {
	_ = ctx
	dsn := dbConfig.DSN()
	conn, err := realsql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	err = RunMigrations(dsn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewApp(dispatcher *service.Dispatcher, adapter in.TgClient, consumer in.UpdatesConsumer,
	cfg *AppConfig, workers []*out.UpdateSender, repo service.ProcessedEventsRepository) *App {
	return &App{
		dispatcher:            dispatcher,
		adapter:               adapter,
		updatesConsumer:       consumer,
		config:                cfg,
		updateHandlingWorkers: workers,
		repo:                  repo,
	}
}

// TODO: многовато логики. надо вынести что то
func (a *App) RunService(ctx context.Context) {
	slog.Info("Starting service. Accepting user messages")

	go func() {
		err := a.updatesConsumer.Run(ctx)
		if err != nil {
			slog.Error("Error starting updates consumer", "error", err)
		}
	}()

	tgUpdates := a.adapter.GetUpdates()

	for _, worker := range a.updateHandlingWorkers {
		go func() {
			worker.Run(ctx)
		}()
	}

	for {
		select {
		case update, ok := <-tgUpdates:
			if !ok {
				tgUpdates = nil
				slog.Info("Telegram Update channel closed")
				break
			}
			err := a.processTgUpdate(ctx, update)
			if err != nil {
				slog.Error("Error processing telegram update", "error", err)
			}

		case <-ctx.Done():
			slog.Info("Shutting down service")
			return
		}
	}
}

func (a *App) processTgUpdate(ctx context.Context, update model.ChatUpdate) error {
	if update.Message == nil {
		return errors.New("telegram update message is nil")
	}

	serverResponse, err := a.dispatcher.Dispatch(ctx, update.Message.Text, update.ChatID)
	if err != nil {
		return fmt.Errorf("dispatching error: %w", err)
	}

	err = a.adapter.SendMessage(update.ChatID, serverResponse)
	if err != nil {
		return fmt.Errorf("TG API sending message error: %w", err)
	}
	return nil
}
