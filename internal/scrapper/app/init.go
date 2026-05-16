package app

import (
	"context"
	realsql "database/sql"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/adapter/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	apiclients "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/api_clients"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"
	grpc2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/grpc"
	http_out "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/httpclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/kafka"
	querybuilder "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository/query_builder"
	rawsql "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository/raw_sql"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

func buildTrackers(config *ScrapperAppConfig) []service.LinkUpdater {
	github := apiclients.NewGithubClient(config.GithubToken, httpclient.NewReliableHTTPClient(config.HTTPReqConfig))
	stackOF := apiclients.NewStackOverflowClient(config.StackOFToken, httpclient.NewReliableHTTPClient(config.HTTPReqConfig))
	res := []service.LinkUpdater{github, stackOF}
	return res
}

func BuildRepos(accessType string, db *realsql.DB) (*out.Repositories, error) {
	switch accessType {
	case "sql":
		chatRepo := rawsql.NewChatRepository(db)
		linkRepo := rawsql.NewLinkRepository(db)
		chatLinkRepo := rawsql.NewChatLinkRepository(db)
		tagRepo := rawsql.NewTagRepository(db)
		linkTagRepo := rawsql.NewLinkTagRepository(db)
		outboxRepo := rawsql.NewOutboxRepository(db)
		return &out.Repositories{
			ChatRepo:     chatRepo,
			LinkRepo:     linkRepo,
			ChatLinkRepo: chatLinkRepo,
			TagRepo:      tagRepo,
			LinkTagRepo:  linkTagRepo,
			OutboxRepo:   outboxRepo,
		}, nil

	case "query":
		psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		chatRepo := querybuilder.NewChatRepository(db, psql)
		linkRepo := querybuilder.NewLinkRepository(db, psql)
		chatLinkRepo := querybuilder.NewChatLinkRepository(db, psql)
		tagRepo := querybuilder.NewTagRepository(db, psql)
		linkTagRepo := querybuilder.NewLinkTagRepository(db, psql)
		outboxRepo := querybuilder.NewOutboxRepository(db, psql)
		return &out.Repositories{
			ChatRepo:     chatRepo,
			LinkRepo:     linkRepo,
			ChatLinkRepo: chatLinkRepo,
			TagRepo:      tagRepo,
			LinkTagRepo:  linkTagRepo,
			OutboxRepo:   outboxRepo,
		}, nil
	default:
		return nil, fmt.Errorf("not supported type of repository: %s", accessType)
	}
}

// TODO: clean up here
func BuildLinkRuntime(ctx context.Context, config ScrapperAppConfig, trackers []service.LinkUpdater,
	trackingService *service.LinkTrackingService) (*LinkUpdatingRuntime, error) {

	slog.Info("building link runtime with config", "config", config.LinkTrackingConfig)
	jobsCap := 200
	jobsChan := make(chan []model.Link, jobsCap)

	var notifier service.MessageSender
	var fallback service.MessageSender
	var err error

	kafkaNotifier := kafka.BuildKafkaProducer(config.KafkaConfig)

	switch config.NotificationType {
	case "http":
		notifier = http_out.NewBotHTTPNotifier(config.BotURL, httpclient.NewReliableHTTPClient(config.HTTPReqConfig))
		fallback = kafkaNotifier
	case "grpc":
		notifier, err = grpc2.NewBotGrpcClient(config.BotURL)
		if err != nil {
			return nil, fmt.Errorf("error building grpc notifier: %w", err)
		}
		fallback = kafkaNotifier
	default:
		notifier = kafkaNotifier
		fallback = nil
	}

	notifierClient := service.NewNotifierClient(notifier, fallback)

	outboxRelay := service.NewOutboxRelay(notifierClient, trackingService)

	linkProcessor := service.NewLinkProcessor(jobsChan, trackingService)

	scheduler, err := service.NewScheduler(trackingService, jobsChan,
		config.LinkTrackingConfig.BatchSize, config.LinkTrackingConfig.LinksOldAge)

	if err != nil {
		return nil, fmt.Errorf("error while building scheduler: %w", err)
	}

	for _, tracker := range trackers {
		linkProcessor.RegisterUpdater(tracker.GetDomain(), tracker)
	}

	return &LinkUpdatingRuntime{
		scheduler:     scheduler,
		outboxRelay:   outboxRelay,
		linkProcessor: linkProcessor,
		cfg:           config.LinkTrackingConfig,
	}, nil
}

func InitCache(cfg cache.Config) (service.LinkCache, error) {
	linkCache, err := cache.NewValkeyClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("init cache client: %w", err)
	}
	return linkCache, nil
}
