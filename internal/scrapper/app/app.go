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

	trackers := buildTrackers(config)
	linkResolver := service.NewLinkResolver(trackers)

	chatService := service.NewChatService(repo, linkResolver)

	//router := http.NewServer(":"+config.Port, chatService)
	grpcServer := grpc.NewScrapperServer(chatService)
	scheduler, err := buildScheduler(*config, trackers, chatService)
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

func buildTrackers(config *ScrapperAppConfig) []service.LinkUpdater {
	github := service.NewGithubClient(config.GithubToken)
	stackOF := service.NewStackOverflowClient(config.StackOFToken)
	res := []service.LinkUpdater{github, stackOF}
	return res
}

func buildScheduler(config ScrapperAppConfig, trackers []service.LinkUpdater, chatService *service.ChatService) (*service.Scheduler, error) {
	notifier := http2.NewBotHttpNotifier(config.BotUrl)

	scheduler, err := service.NewScheduler(notifier, chatService)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler %v", err)
	}
	for _, tracker := range trackers {
		scheduler.RegisterUpdater(tracker.GetDomain(), tracker)
	}

	return scheduler, nil
}
