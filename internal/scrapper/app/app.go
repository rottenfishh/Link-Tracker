package app

import (
	"context"
	realsql "database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/grpc"
	http_in "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/httpserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/middleware"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/cache"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type App struct {
	linkRuntime *LinkUpdatingRuntime
	server      *http_in.Server
	grpcServer  *grpc.ScrapperServer
}

type LinkUpdatingRuntime struct {
	scheduler     *service.Scheduler
	outboxRelay   *service.OutboxRelay
	linkProcessor *service.LinkProcessor
	cfg           LinkTrackingConfig
}

func BuildApp(ctx context.Context, config *ScrapperAppConfig) (*App, error) {
	slog.Info("creating scrapper app", "database_url", config.DatabaseConfig.DatabaseName, "access_type", config.DatabaseConfig.AccessType)
	pool, err := connectDB(context.Background(), config.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to database: %w", err)
	}

	repos, err := BuildRepos(config.DatabaseConfig.AccessType, pool)
	if err != nil {
		return nil, fmt.Errorf("error while building repos: %w", err)
	}

	trackers := buildTrackers(config)
	linkResolver := service.NewLinkResolver(trackers)

	txManager := repository.NewTransactionManager(pool)

	linkCache, err := InitCache(config.CacheConfig)
	if err != nil {
		slog.Error("error while initializing cache client", "error", err)
		slog.Info("creating dummy no op cache client")
		linkCache = &cache.NoOpCache{}
	}

	chatService := service.NewChatService(*repos, linkResolver, txManager, linkCache)

	trackingService := service.NewLinkTrackingService(repos, txManager)

	rateLimiter := middleware.NewRateLimiter(config.RateLimitConfig)
	grpcServer := grpc.NewScrapperServer(chatService, rateLimiter, config.Port, config.GatewayPort)

	linkRuntime, err := BuildLinkRuntime(ctx, *config, trackers, trackingService)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler: %w", err)
	}

	return &App{linkRuntime: linkRuntime, grpcServer: grpcServer}, nil
}

func NewApp(linkRuntime *LinkUpdatingRuntime, server *http_in.Server,
	grpcServer *grpc.ScrapperServer) *App {
	return &App{
		linkRuntime: linkRuntime,
		server:      server,
		grpcServer:  grpcServer,
	}
}

func (a *App) Run(ctx context.Context) error {
	err := a.linkRuntime.Start(ctx)
	if err != nil {
		return fmt.Errorf("error starting scheduler: %w", err)
	}

	err = a.grpcServer.RunServer(ctx)
	if err != nil {
		return fmt.Errorf("error starting router: %w", err)
	}

	return nil
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

func (r *LinkUpdatingRuntime) Start(ctx context.Context) error {
	err := r.scheduler.StartScheduler(ctx, r.cfg.SchedulerInterval)
	if err != nil {
		return fmt.Errorf("error while starting scheduler: %w", err)
	}

	for i := range r.cfg.NThreads {
		slog.Info("Starting worker thread", "thread", i)
		go r.linkProcessor.Start(ctx)
	}
	go func() {
		err = r.outboxRelay.Run(ctx)
	}()
	if err != nil {
		return fmt.Errorf("error while starting outboxRelay: %w", err)
	}

	return nil
}
