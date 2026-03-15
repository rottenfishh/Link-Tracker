package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
)

// TODO: fix this. interface for updaters
// have type of link on link and type of notifier. go over links and call needed notifier from map
// map [site model] = notifier
type Scheduler struct {
	gocron.Scheduler
	service  *ChatService
	notifier out.Notifier
	updaters map[string]LinkUpdater
}

func NewScheduler(notifier out.Notifier, service *ChatService) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	updaters := make(map[string]LinkUpdater)

	return &Scheduler{s, service, notifier, updaters}, nil
}

func (s *Scheduler) RegisterUpdater(name string, updater LinkUpdater) {
	s.updaters[name] = updater
}

func (s *Scheduler) GetUpdater(name string) LinkUpdater {
	return s.updaters[name]
}

// TODO: iterate over links first
// TODO: нужно идти по линкам, а не по чатам, и уведомлять всех, кто подписан на линк.
func (s *Scheduler) updateLinks(ctx context.Context) error {
	links, err := s.service.GetLinks(ctx)
	if err != nil {
		return err
	}

	for _, link := range links {
		tracker := s.updaters[link.Domain]
		if tracker == nil {
			slog.Error("No link tracker for this url found: ", "url", link)
			continue
		}

		slog.Info("Requesting update from ", "link", link.Link)
		update, err := tracker.GetUpdates(link.Link)
		if err != nil {
			return err
		}

		if update.LastModified.After(link.LastUpdated) {
			err = s.updateUsers(ctx, &link, update)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scheduler) updateUsers(ctx context.Context, link *model.Link, update *model.Update) error {
	subscribers, err := s.service.GetSubscribersByLink(ctx, link)
	if err != nil {
		return err
	}

	slog.Info("Updating ", link.Link, "time", link.LastUpdated.String(), " to ", update.LastModified)
	upd := model.NewLinkUpdate(1, link.Link, "New event from given link", subscribers)

	err = s.notifier.SendUpdate(context.Background(), *upd)
	if err != nil {
		slog.Error("error sending update to bot", "error", err)
		return err
	}

	link.LastUpdated = update.LastModified
	_, err = s.service.UpdateLink(ctx, link)
	if err != nil {
		slog.Error("error updating link time in chatRepo", "error", err)
		return err
	}
	return nil
}

func (s *Scheduler) StartScheduler() error {
	j, err := s.NewJob(
		gocron.DurationJob(
			30*time.Second,
		),
		gocron.NewTask(s.updateLinks),
		gocron.WithSingletonMode(gocron.LimitModeWait),
	)
	if err != nil {
		return fmt.Errorf("error creating scheduler job: %v", err)
	}
	fmt.Println(j.ID())

	s.Start()
	return nil
}
