package app

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/application"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure/in"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure/out"
)

type App struct {
	scheduler *application.Scheduler
	router    *gin.Engine
	config    AppConfig
}

func NewApp() (*App, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error while loading config for scheduler %v", err)
	}

	repo := out.NewInMemoryRepo()

	router := buildServer(repo)

	scheduler, err := buildScheduler(*config, repo)
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
	err = a.router.Run(":" + a.config.Port)
	if err != nil {
		return fmt.Errorf("error starting router %v", err)
	}
	return nil
}

func buildScheduler(config AppConfig, repo out.ChatRepository) (*application.Scheduler, error) {
	github := application.NewGithubClient(config.GithubToken)
	stackOF := application.NewStackOverflowClient(config.StackOFToken)
	notifier := infrastructure.NewBotHttpNotifier(config.BotUrl)

	scheduler, err := application.NewScheduler(notifier, repo)
	if err != nil {
		return nil, fmt.Errorf("error while building scheduler %v", err)
	}
	scheduler.RegisterUpdater("github", github)
	scheduler.RegisterUpdater("stackof", stackOF)

	return scheduler, nil
}

func buildServer(repo out.ChatRepository) *gin.Engine {
	handler := in.NewHttpHandler(repo)
	server := gin.Default()
	registerRoutes(server, handler)
	return server
}

func registerRoutes(router *gin.Engine, handler *in.HttpHandler) {
	router.POST("/tg-chat/{id}", handler.RegisterChat)
	router.DELETE("/tg-chat/{id}", handler.DeleteChat)
	router.GET("/links/{id}", handler.GetLinksByChatId)
	router.POST("links/{id}", handler.AddLink)
	router.DELETE("links/{id}", handler.DeleteLink)
}
