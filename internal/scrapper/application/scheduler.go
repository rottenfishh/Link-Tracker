package application

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/infrastructure/out"
)

// TODO: fix this. interface for updaters
type Scheduler struct {
	gocron.Scheduler
	notifier     infrastructure.Notifier
	repo         out.InMemoryRepo
	githubClient *GithubClient
	stackClient  *StackOverflowClient
}

func NewScheduler(githubClient *GithubClient, stackClient *StackOverflowClient, notifier infrastructure.Notifier) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	repo := out.InMemoryRepo{make(map[int64]*domain.Chat), make(map[string][]int64)}
	return &Scheduler{s, notifier, repo, githubClient, stackClient}, nil
}

// TODO: нужно идти по линкам, а не по чатам, и уведомлять всех, кто подписан на линк.
func (s *Scheduler) updateUsers() error {
	for _, chat := range s.repo.Chats {
		err := s.updateLinks(chat)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) updateLinks(chat *domain.Chat) error {
	for _, link := range chat.Links {
		update, err := s.githubClient.GetUpdates(link.Link)
		if err != nil {
			return err
		}

		if update.LastModified.After(link.LastModified) {
			chats := []int64{chat.Id}
			upd := domain.LinkUpdate{1, link.Link, "New event from given link", chats}
			err = s.notifier.SendUpdate(upd)
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
