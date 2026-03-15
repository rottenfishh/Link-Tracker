package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/http"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository/sql"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type App struct {
	scheduler *service.Scheduler
	//server     *http.Server
	grpcServer *grpc.ScrapperServer
	config     *ScrapperAppConfig
	dbConfig   *DatabaseConfig
	repos      Repositories
}

type Repositories struct {
	ChatRepo     out.ChatRepository
	LinkRepo     out.LinkRepository
	ChatLinkRepo out.ChatLinkRepository
}

func NewApp(config *ScrapperAppConfig) (*App, error) {
	slog.Info(config.DatabaseConfig.DatabaseUrl, config.DatabaseConfig.AccessType)
	pool, err := connectDB(context.Background(), config.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to database %v", err)
	}

	repos, err := buildRepos(config.DatabaseConfig.AccessType, pool)
	if err != nil {
		return nil, fmt.Errorf("error while building repos %v", err)
	}

	chatService := service.NewChatService(repos.ChatRepo, repos.LinkRepo, repos.ChatLinkRepo)

	//httpRouter := http.NewServer(":"+config.Port, chatService)
	grpcServer := grpc.NewScrapperServer(chatService)
	scheduler, err := buildScheduler(*config, chatService)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler %v", err)
	}

	return &App{scheduler: scheduler, grpcServer: grpcServer, config: config}, nil
}

// TODO: add gateway port
func (a *App) Run() error {
	err := a.scheduler.StartScheduler()
	if err != nil {
		return fmt.Errorf("error starting scheduler %v", err)
	}

	slog.Info("Starting scrapper service at port " + a.config.Port)
	err = a.grpcServer.RunServer("8089", a.config.Port)
	if err != nil {
		return fmt.Errorf("error starting router %v", err)
	}
	return nil
}

func buildScheduler(config ScrapperAppConfig, chatService *service.ChatService) (*service.Scheduler, error) {
	github := service.NewGithubClient(config.GithubToken)
	stackOF := service.NewStackOverflowClient(config.StackOFToken)
	notifier := http.NewBotHttpNotifier(config.BotUrl)

	scheduler, err := service.NewScheduler(notifier, chatService)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler %v", err)
	}
	scheduler.RegisterUpdater("github", github)
	scheduler.RegisterUpdater("stackof", stackOF)

	return scheduler, nil
}

func connectDB(ctx context.Context, dbConfig *DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := dbConfig.DSN()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database %v", err)
	}
	err = RunMigrations(dsn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func buildRepos(accessType string, pool *pgxpool.Pool) (*Repositories, error) {
	switch accessType {
	case "sql":
		chatRepo := sql.NewChatRepository(pool)
		linkRepo := sql.NewLinkRepository(pool)
		chatLinkRepo := sql.NewChatLinkRepository(pool)
		return &Repositories{chatRepo, linkRepo, chatLinkRepo}, nil
	case "query":
		//TODO: implement me
	default:
		return nil, fmt.Errorf("not supported type of repository: " + accessType)
	}
	return nil, nil
}
