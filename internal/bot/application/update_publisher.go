package application

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/domain"
)

type UpdatePublisher struct {
	updates chan domain.LinkUpdate
}

func NewUpdatePublisher() *UpdatePublisher {
	return &UpdatePublisher{updates: make(chan domain.LinkUpdate)}
}

func (p *UpdatePublisher) PublishUpdate(update domain.LinkUpdate) {
	p.updates <- update
}

func (p *UpdatePublisher) GetUpdates() chan domain.LinkUpdate {
	return p.updates
}
