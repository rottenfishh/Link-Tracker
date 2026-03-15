package app

import (
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/http"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	http2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/http"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type App struct {
	scheduler  *service.Scheduler
	server     *http.Server
	grpcServer *grpc.ScrapperServer
	config     *ScrapperAppConfig
}

func NewApp(config *ScrapperAppConfig) (*App, error) {

	repo := out.NewInMemoryRepo()
	chatService := service.NewChatService(repo)

	//router := http.NewServer(":"+config.Port, chatService)
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
	notifier := http2.NewBotHttpNotifier(config.BotUrl)

	scheduler, err := service.NewScheduler(notifier, chatService)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler %v", err)
	}
	scheduler.RegisterUpdater("github", github)
	scheduler.RegisterUpdater("stackof", stackOF)

	return scheduler, nil
}
