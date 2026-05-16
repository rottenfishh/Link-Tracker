package out

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/adapter/in"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
)

type UpdateSender struct {
	Updates chan service.UpdateJob
	Reports chan service.ReportJob
	adapter in.TgClient
}

func NewUpdatesSender(publisher *service.UpdatePublisher, adapter in.TgClient) *UpdateSender {
	updates := publisher.GetUpdatesChan()
	reports := publisher.GetReportsChan()
	return &UpdateSender{Updates: updates, Reports: reports, adapter: adapter}
}

func (s *UpdateSender) Run(ctx context.Context) {
	slog.Info("UpdatesSender started")
	for {
		select {
		case update, ok := <-s.Updates:
			if !ok {
				s.Updates = nil
				slog.Info("Update channel closed")
				continue
			}
			err := s.processLinkUpdate(ctx, update)
			update.Result <- err
			if err != nil {
				slog.Error("error processing update", "update", update, "err", err)
			}
		case report, ok := <-s.Reports:
			if !ok {
				s.Reports = nil
				slog.Info("Report channel closed")
				continue
			}
			err := s.processReportUpdate(ctx, report)
			report.Result <- err
			if err != nil {
				slog.Error("error processing report", "report", report, "err", err)
			}
		case <-ctx.Done():
			return
		}

		if s.Reports == nil && s.Updates == nil {
			return
		}
	}
}

func (s *UpdateSender) processLinkUpdate(ctx context.Context, update service.UpdateJob) error {
	_ = ctx
	newMsg := service.BuildUpdate(update.LinkUpdate)

	slog.Debug("New Link update received in bot", "newMsg", newMsg)
	for _, id := range update.TgChatIDs {
		err := s.adapter.SendMessage(id, newMsg)
		if err != nil {
			return fmt.Errorf("error sending message to telegram user %d: %w", id, err)
		}
	}
	return nil
}

func (s *UpdateSender) processReportUpdate(ctx context.Context, report service.ReportJob) error {
	_ = ctx
	newMsg := service.BuildReport(report.Report)
	slog.Info("App received report", "message", newMsg)
	err := s.adapter.SendMessage(report.ChatID, newMsg)

	if err != nil {
		return fmt.Errorf("error sending message to telegram user %d: %w", report.ChatID, err)
	}
	return nil
}
