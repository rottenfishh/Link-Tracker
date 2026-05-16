package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type LinkTrackingService struct {
	r         out.Repositories
	TxManager *repository.TransactionManager
}

func NewLinkTrackingService(r *out.Repositories, txManager *repository.TransactionManager) *LinkTrackingService {
	return &LinkTrackingService{
		r:         *r,
		TxManager: txManager,
	}
}

func (s *LinkTrackingService) GetSubscribersByLinkID(ctx context.Context, linkID int64) ([]int64, error) {
	subscribers, err := s.r.ChatLinkRepo.GetChatIDsByLinkID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("getting subscribers by link ID: %w", err)
	}

	return subscribers, nil
}

func (s *LinkTrackingService) UpdateLinkLastModifiedOutbox(ctx context.Context, link *model.Link,
	update *model.LinkUpdate) (*model.Link, error) {
	var updatedLink *model.Link
	var err error

	err = s.TxManager.WithTx(ctx, func(ctx context.Context) error {
		updatedLink, err = s.r.LinkRepo.UpdateLink(ctx, link.ID, link)
		if err != nil {
			return fmt.Errorf("updating link: %w", err)
		}

		var outbox *model.Outbox
		outbox, err = buildOutboxFromUpdate(update)
		if err != nil {
			return fmt.Errorf("creating outbox struct from update: %w", err)
		}

		_, err = s.r.OutboxRepo.Save(ctx, outbox)
		if err != nil {
			return fmt.Errorf("saving outbox: %w", err)
		}
		return nil
	})
	return updatedLink, nil
}

func (s *LinkTrackingService) SaveReportOutbox(ctx context.Context, report *model.Report) (*model.Report, error) {
	outbox, err := buildOutboxFromReport(report)
	if err != nil {
		return nil, fmt.Errorf("creating outbox struct from report: %w", err)
	}
	_, err = s.r.OutboxRepo.Save(ctx, outbox)
	if err != nil {
		return nil, fmt.Errorf("saving outbox: %w", err)
	}
	return report, nil
}

func (s *LinkTrackingService) GetLinksOlderThanWithLimitAndOffset(ctx context.Context, time time.Time, limit, offset int) ([]model.Link, error) {
	links, err := s.r.LinkRepo.GetLinksOlderThan(ctx, time, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("getting old links: %w", err)
	}

	return links, nil
}

func (s *LinkTrackingService) ForEachLinkOlderThan(ctx context.Context, time time.Time, limit int, fn func(ctx context.Context, link *model.Link) error) error {
	offset := 0
	for {
		links, err := s.r.LinkRepo.GetLinksOlderThan(ctx, time, limit, offset)
		if err != nil {
			slog.Error("GetLinksOlderThan error", "error", err)
			return fmt.Errorf("GetLinksOlderThan: %w", err)
		}
		if len(links) == 0 {
			slog.Debug("no links for update")
			break
		}
		for _, link := range links {
			if err = fn(ctx, &link); err != nil {
				slog.Error(err.Error())
				continue
			}
		}
		offset += len(links)
	}
	return nil
}

func (s *LinkTrackingService) GetPendingOutbox(ctx context.Context) ([]model.Outbox, error) {
	outbox, err := s.r.OutboxRepo.GetPendingOutbox(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting pending outbox: %w", err)
	}
	return outbox, nil
}

func (s *LinkTrackingService) MarkOutboxDone(ctx context.Context, id string, timeProcessed time.Time) (*model.Outbox, error) {
	outbox, err := s.r.OutboxRepo.UpdateProcessedAt(ctx, id, timeProcessed)
	if err != nil {
		return nil, fmt.Errorf("marking outbox as done: %w", err)
	}
	return outbox, nil
}

func buildOutboxFromUpdate(upd *model.LinkUpdate) (*model.Outbox, error) {
	msgJSON, err := json.Marshal(upd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal outbox update: %w", err)
	}

	return &model.Outbox{
		ID:          upd.ID,
		Topic:       "updates",
		MessageType: "link-update",
		Payload:     (msgJSON),
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func buildOutboxFromReport(report *model.Report) (*model.Outbox, error) {
	msgJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal outbox report: %w", err)
	}

	return &model.Outbox{
		ID:          report.ID,
		Topic:       "reports",
		MessageType: "report",
		Payload:     (msgJSON),
		CreatedAt:   time.Now().UTC(),
	}, nil
}
