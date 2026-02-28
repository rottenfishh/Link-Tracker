package application

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure/out"
)

// TODO: fix this. interface for updaters
// have type of link on link and type of notifier. go over links and call needed notifier from map
// map [site domain] = notifier
type Scheduler struct {
	gocron.Scheduler
	repo     out.ChatRepository
	notifier infrastructure.Notifier
	updaters map[string]LinkUpdater
}

func NewScheduler(notifier infrastructure.Notifier, repo out.ChatRepository) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	updaters := make(map[string]LinkUpdater)

	//repo := out.InMemoryRepo{make(map[int64]*domain.Chat), make(map[string][]int64)}
	return &Scheduler{s, repo, notifier, updaters}, nil
}

func (s *Scheduler) RegisterUpdater(name string, updater LinkUpdater) {
	s.updaters[name] = updater
}

func (s *Scheduler) GetUpdater(name string) LinkUpdater {
	return s.updaters[name]
}

// TODO: нужно идти по линкам, а не по чатам, и уведомлять всех, кто подписан на линк.
func (s *Scheduler) updateUsers() error {
	chats, err := s.repo.GetChats()
	if err != nil {
		return err
	}
	for _, chat := range chats {
		err = s.updateLinks(chat)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) updateLinks(chat domain.Chat) error {
	for _, link := range chat.Links {
		tracker := s.updaters[link.Domain]
		if tracker == nil {
			slog.Error("No link tracker for this url found: ", "url", link)
			continue
		}

		update, err := tracker.GetUpdates(link.Link)
		if err != nil {
			return err
		}

		if update.LastModified.After(link.LastModified) {
			chats := []int64{chat.Id}
			upd := domain.NewLinkUpdate(1, link.Link, "New event from given link", chats)
			err = s.notifier.SendUpdate(*upd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scheduler) StartScheduler() error {
	j, err := s.NewJob(
		gocron.DurationJob(
			100*time.Second,
		),
		gocron.NewTask(s.updateUsers()),
	)
	if err != nil {
		return fmt.Errorf("error creating scheduler job: %v", err)
	}
	fmt.Println(j.ID())

	s.Start()
	return nil
}
