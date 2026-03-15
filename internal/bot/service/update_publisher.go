package service

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type UpdatePublisher struct {
	updates chan model.LinkUpdate
}

func NewUpdatePublisher() *UpdatePublisher {
	return &UpdatePublisher{updates: make(chan model.LinkUpdate, 100)}
}

func (p *UpdatePublisher) PublishUpdate(update model.LinkUpdate) {
	p.updates <- update
}

func (p *UpdatePublisher) GetUpdates() chan model.LinkUpdate {
	return p.updates
}
