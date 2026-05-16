package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

// LinkProcessor получает батч ссылок, получает по ним апдейты и отдает их в канал нотифаеру.
type LinkProcessor struct {
	Trackers        map[string]LinkUpdater
	TrackingService *LinkTrackingService
	JobsChan        chan []model.Link
}

func NewLinkProcessor(jobsChan chan []model.Link, trackingService *LinkTrackingService) *LinkProcessor {
	trackers := make(map[string]LinkUpdater)

	return &LinkProcessor{Trackers: trackers, TrackingService: trackingService, JobsChan: jobsChan}
}

func (p *LinkProcessor) RegisterUpdater(name string, updater LinkUpdater) {
	p.Trackers[name] = updater
}

func (p *LinkProcessor) SendJobs(links []model.Link) {
	p.JobsChan <- links
}

func (p *LinkProcessor) Start(ctx context.Context) {
	slog.Debug("worker started")

	for {
		select {
		case links := <-p.JobsChan:
			p.ProcessLinks(ctx, links)
		case <-ctx.Done():
			slog.Info("link processor thread cancelled")
			return
		}
	}
}

func (p *LinkProcessor) ProcessLinks(ctx context.Context, links []model.Link) {
	slog.Debug("Updating links")

	var failedLinks []model.LinkUpdateError
	for _, link := range links {
		err := p.Process(ctx, &link)
		if err != nil {
			slog.Error("updating link failed", "link", link.Link, "error", err)

			failedLink := model.LinkUpdateError{Link: link.Link, Err: err.Error(),
				Message: "failed to update link", LinkID: link.ID}
			failedLinks = append(failedLinks, failedLink)
		}
	}

	if len(failedLinks) > 0 {
		err := p.SendFailedLinks(ctx, failedLinks)
		if err != nil {
			slog.Error("sending report failed", "report", failedLinks, "error", err)
		}
	}
}

func (p *LinkProcessor) Process(ctx context.Context, link *model.Link) error {
	tracker := p.Trackers[link.Domain]
	if tracker == nil {
		slog.Error("tracker not found for domain", "domain", link.Domain)
		return fmt.Errorf("no link tracker for this url found %s", link.Link)
	}

	slog.Debug("Requesting update from ", "link", link.Link)
	updates, err := tracker.GetUpdates(link)
	if err != nil {
		return fmt.Errorf("error getting update for %s: %w", link.Link, err)
	}

	for _, update := range updates {
		err = p.Notify(ctx, link, update)
		if err != nil {
			return fmt.Errorf("error updating users for %s: %w", link.Link, err)
		}
	}
	return nil
}

func (p *LinkProcessor) Notify(ctx context.Context, link *model.Link, update model.Update) error {
	subscribers, err := p.TrackingService.GetSubscribersByLinkID(ctx, link.ID)
	if err != nil {
		return err
	}

	slog.Debug("updating subscribers for link", "link", link.Link, "from", link.LastUpdated.String(), "to", update.LastModified)
	upd := model.NewLinkUpdate(uuid.New().String(), link.Link, update, subscribers)

	link.LastUpdated = update.LastModified
	_, err = p.TrackingService.UpdateLinkLastModifiedOutbox(ctx, link, upd)
	if err != nil {
		slog.Error("error updating link time in linkRepo", "error", err)
		return err
	}
	return nil
}

func (p *LinkProcessor) SendFailedLinks(ctx context.Context, links []model.LinkUpdateError) error {
	reports := make(map[int64]*model.Report)
	if len(links) == 0 {
		slog.Debug("no links to send")
		return nil
	}

	for _, link := range links {
		subscribers, err := p.TrackingService.GetSubscribersByLinkID(ctx, link.LinkID)
		if err != nil {
			return err
		}

		for _, subscriber := range subscribers {
			if _, ok := reports[subscriber]; !ok {
				reports[subscriber] = model.NewReport(uuid.New().String(), subscriber)
			}
			reports[subscriber].LinkUpdateErrors = append(reports[subscriber].LinkUpdateErrors, &link)
		}
	}
	if len(reports) == 0 {
		slog.Debug("no reports to send")
		return nil
	}
	for _, report := range reports {
		slog.Debug("sent report", "report", report)

		_, err := p.TrackingService.SaveReportOutbox(ctx, report)
		if err != nil {
			slog.Error("error saving report", "error", err)
			return err
		}
	}
	return nil
}
