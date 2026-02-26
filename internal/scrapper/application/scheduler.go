package application

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type Scheduler struct {
	gocron.Scheduler
	githubClient *GithubClient
	stackClient  *StackOverflowClient
}

func NewScheduler(githubClient *GithubClient, stackClient *StackOverflowClient) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &Scheduler{s, githubClient, stackClient}, nil
}

func (s *Scheduler) StartScheduler() error {

	j, err := s.NewJob(
		gocron.DurationJob(
			200*time.Second,
		),
		gocron.NewTask(s.githubClient.GetUpdates("blob")),
	)
	if err != nil {
		return fmt.Errorf("error creating github job: %v", err)
	}

	j, err = s.NewJob(
		gocron.DurationJob(
			200*time.Second,
		),
		gocron.NewTask(s.stackClient.GetUpdates("blob")),
	)
	if err != nil {
		return fmt.Errorf("error creating stack overflow job: %v", err)
	}
	fmt.Println(j.ID())

	s.Start()
	return nil
}
