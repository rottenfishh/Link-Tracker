package app

import (
	"fmt"
	"log/slog"

	http2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/in/http"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/service"
)

type App struct {
	scheduler *service.Scheduler
	server    *http2.Server
	config    ScrapperAppConfig
}

func NewApp() (*App, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error while loading config for scheduler %v", err)
	}

	repo := out.NewInMemoryRepo()
	chatService := service.NewChatService(repo)

	router := http2.NewServer(":"+config.Port, chatService)

	scheduler, err := buildScheduler(*config, chatService)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler %v", err)
	}

	return &App{scheduler, router, *config}, nil
}

func (a *App) Run() error {
	err := a.scheduler.StartScheduler()
	if err != nil {
		return fmt.Errorf("error starting scheduler %v", err)
	}

	slog.Info("Starting scrapper service at port " + a.config.Port)
	err = a.server.Run()
	if err != nil {
		return fmt.Errorf("error starting router %v", err)
	}
	return nil
}

func buildScheduler(config ScrapperAppConfig, chatService *service.ChatService) (*service.Scheduler, error) {
	github := service.NewGithubClient(config.GithubToken)
	stackOF := service.NewStackOverflowClient(config.StackOFToken)
	notifier := out.NewBotHttpNotifier(config.BotUrl)

	scheduler, err := service.NewScheduler(notifier, chatService)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler %v", err)
	}
	scheduler.RegisterUpdater("github", github)
	scheduler.RegisterUpdater("stackof", stackOF)

	return scheduler, nil
}
