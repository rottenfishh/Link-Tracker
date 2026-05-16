package service

import (
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

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
			link.Title, err = tracker.GetTitle(link.Link)
			if err != nil {
				return nil, fmt.Errorf("getting title for %s link: %w", link.Link, err)
			}

			return link, nil
		}
	}
	return nil, model.ErrInvalidRequest
}
