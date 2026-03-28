package service

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"

type LinkResolver struct {
	Trackers []LinkUpdater
}

func NewLinkResolver(trackers []LinkUpdater) *LinkResolver {
	return &LinkResolver{trackers}
}

func (r *LinkResolver) FormatLink(link *model.Link) (*model.Link, error) {
	for _, tracker := range r.Trackers {
		if linkResult, err := tracker.FormatLink(link.Link); err == nil {
			link.FormattedLink = linkResult
			link.Domain = tracker.GetDomain()
			return link, nil
		}
	}
	return nil, model.ErrInvalidRequest
}
