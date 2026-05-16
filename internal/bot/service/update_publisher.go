package service

import (
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type UpdatePublisher struct {
	updates chan UpdateJob
	reports chan ReportJob
}

// TODO: get from config
func NewUpdatePublisher(updatesCap, reportsCap int) *UpdatePublisher {
	return &UpdatePublisher{updates: make(chan UpdateJob, updatesCap),
		reports: make(chan ReportJob, reportsCap)}
}

func (p *UpdatePublisher) PublishUpdate(update model.LinkUpdate) error {
	resultChan := make(chan error, 1)

	updateJob := UpdateJob{
		LinkUpdate: update,
		Result:     resultChan,
	}

	p.updates <- updateJob
	err := <-resultChan

	if err != nil {
		return fmt.Errorf("error sending update in publisher: %w", err)
	}

	return nil
}

func (p *UpdatePublisher) PublishReport(report model.Report) error {
	resultChan := make(chan error, 1)

	reportJob := ReportJob{
		Report: report,
		Result: resultChan,
	}

	p.reports <- reportJob
	err := <-resultChan

	if err != nil {
		return fmt.Errorf("error sending report in publisher: %w", err)
	}

	return nil
}

func (p *UpdatePublisher) PublishUpdateNoWait(update model.LinkUpdate) {
	resultChan := make(chan error, 1)

	updateJob := UpdateJob{
		LinkUpdate: update,
		Result:     resultChan,
	}

	p.updates <- updateJob
}

func (p *UpdatePublisher) PublishReportNoWait(report model.Report) {
	resultChan := make(chan error, 1)

	reportJob := ReportJob{
		Report: report,
		Result: resultChan,
	}

	p.reports <- reportJob
}

func (p *UpdatePublisher) GetUpdatesChan() chan UpdateJob {
	return p.updates
}

func (p *UpdatePublisher) GetReportsChan() chan ReportJob {
	return p.reports
}

func (p *UpdatePublisher) ShutDown() {
	close(p.updates)
	close(p.reports)
}
