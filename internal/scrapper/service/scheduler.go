package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Scheduler struct {
	gocron.Scheduler
	trackingService *LinkTrackingService
	jobsChan        chan []model.Link
	linksBatchSize  int
	linksOldAge     int
}

func NewScheduler(trackingService *LinkTrackingService, jobsChan chan []model.Link, linksBatchSize,
	linksOldAge int) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("creating scheduler: %w", err)
	}

	return &Scheduler{
		Scheduler:       s,
		trackingService: trackingService,
		jobsChan:        jobsChan,
		linksBatchSize:  linksBatchSize,
		linksOldAge:     linksOldAge}, nil
}

func (s *Scheduler) StartScheduler(ctx context.Context, interval int) error {
	j, err := s.NewJob(
		gocron.DurationJob(
			time.Duration(interval)*time.Second,
		),
		gocron.NewTask(s.updateLinks),
		gocron.WithSingletonMode(gocron.LimitModeWait),
	)
	if err != nil {
		return fmt.Errorf("error creating scheduler job: %w", err)
	}
	fmt.Println(j.ID())

	s.Start()

	go func() {
		<-ctx.Done()
		slog.Info("scheduler shutting down")
		err = s.Shutdown()
		if err != nil {
			slog.Error("error shutting down scheduler", "error", err)
		}
	}()

	return nil
}

func (s *Scheduler) updateLinks(ctx context.Context) {
	slog.Debug("scheduler checking links")
	offset := 0
	limit := s.linksBatchSize

	ttl := -time.Duration(s.linksOldAge) * time.Second
	for {
		links, err := s.trackingService.GetLinksOlderThanWithLimitAndOffset(ctx, time.Now().Add(ttl), limit, offset)
		if err != nil || len(links) == 0 {
			slog.Debug("no links found")
			return
		}
		s.jobsChan <- links
		offset += len(links)
	}
}
