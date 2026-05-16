package app

import (
	"context"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in"
	grpc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/grpc"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/httpserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in/kafka"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/commands"
)

func BuildDispatcher(scrapperClient commands.ScrapperClient) *service.Dispatcher {
	help := &commands.HelpCommand{}
	start := commands.NewStartCommand(scrapperClient)
	fallback := &commands.FallBackCommand{}
	track := commands.NewTrackCommand(scrapperClient)
	untrack := commands.NewUntrackCommand(scrapperClient)
	list := commands.NewListCommand(scrapperClient)
	cancel := &commands.CancelCommand{}

	cmds := []service.Command{help, start, fallback, track, untrack, list, cancel}

	d := service.NewDispatcher()
	for _, cmd := range cmds {
		d.Register(cmd)
	}

	return d
}

func BuildConsumer(ctx context.Context, cfg AppConfig, publisher *service.UpdatePublisher,
	repo service.ProcessedEventsRepository) (in.UpdatesConsumer, error) {

	var consumer in.UpdatesConsumer

	switch cfg.NotificationType {
	case "grpc":
		consumer = grpc.NewBotServiceServer(publisher, strconv.Itoa(cfg.Port), strconv.Itoa(cfg.GatewayPort))
	case "http":
		consumer = httpserver.NewServer(":"+strconv.Itoa(cfg.Port), publisher)
	default:
		consumer = kafka.BuildKafkaConsumer(cfg.KafkaConfig, publisher, repo)
	}

	return consumer, nil

}
func BuildUpdateWorkerPool(cfg AppConfig, publisher *service.UpdatePublisher, client in.TgClient) []*out.UpdateSender {
	var workers []*out.UpdateSender

	for range cfg.NThreads {
		worker := out.NewUpdatesSender(publisher, client)
		workers = append(workers, worker)
	}

	return workers
}
