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

// TODO: нужно идти по линкам, а не по чатам, и уведомлять всех, кто подписан на линк.
func (s *Scheduler) updateUsers() error {
	chats, err := s.service.GetChats()
	if err != nil {
		return err
	}
	for _, chat := range chats {
		err = s.updateLinks(chat)
		if err != nil {
			slog.Error("error updating links: ", "error ", err, " chat", chat)
			return err
		}
	}
	return nil
}

// TODO: save new time for link in repo properly
func (s *Scheduler) updateLinks(chat model.Chat) error {
	for _, link := range chat.Links {
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

		if update.LastModified.After(link.LastModified) {
			slog.Info("Updating ", link.Link, "time", link.LastModified, " to ", update.LastModified)
			chats := []int64{chat.Id}
			upd := model.NewLinkUpdate(1, link.Link, "New event from given link", chats)
			err = s.notifier.SendUpdate(context.Background(), *upd)
			if err != nil {
				slog.Error("error sending update to bot", "error", err)
				return err
			}

			link.LastModified = update.LastModified
			_, err = s.service.UpdateLink(chat.Id, link)
			if err != nil {
				slog.Error("error updating link time in repo", "error", err)
				return err
			}
		}
	}
	return nil
}

func (s *Scheduler) StartScheduler() error {
	j, err := s.NewJob(
		gocron.DurationJob(
			30*time.Second,
		),
		gocron.NewTask(s.updateUsers),
		gocron.WithSingletonMode(gocron.LimitModeWait),
	)
	if err != nil {
		return fmt.Errorf("error creating scheduler job: %v", err)
	}
	fmt.Println(j.ID())

	s.Start()
	return nil
}
