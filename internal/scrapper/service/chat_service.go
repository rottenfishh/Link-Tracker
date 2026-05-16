//nolint:mnd // default pagination limit is an intentional service-level constant
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/scrapper/adapter/out/repository"
)

type ChatService struct {
	r            out.Repositories
	linkResolver *LinkResolver
	TxManager    *repository.TransactionManager
	cache        LinkCache
}

func NewChatService(repos out.Repositories,
	linkResolver *LinkResolver, txManager *repository.TransactionManager, cache LinkCache) *ChatService {
	return &ChatService{
		r:            repos,
		linkResolver: linkResolver,
		TxManager:    txManager,
		cache:        cache,
	}
}

func (s *ChatService) RegisterChat(ctx context.Context, id int64) (*model.Chat, error) {
	chat := model.NewChat(id)
	err := s.r.ChatRepo.SaveChat(ctx, chat)
	if err != nil {
		slog.Error("saving chat error", "error", err)
		return nil, fmt.Errorf("saving chat: %w", err)
	}
	return chat, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, id int64) error {
	_, err := s.r.ChatRepo.DeleteChat(ctx, id)
	if err != nil {
		slog.Error("deleting chat", "error", err)
		return fmt.Errorf("deleting chat: %w", err)
	}
	return nil
}

func (s *ChatService) GetChats(ctx context.Context) ([]model.Chat, error) {
	chats, err := s.r.ChatRepo.GetChats(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting chats: %w", err)
	}

	return chats, nil
}

func (s *ChatService) GetLinksByChatID(ctx context.Context, id int64) ([]model.Link, error) {
	links, err := s.cache.Get(ctx, id)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		slog.Error("getting links from cache by chat ID error", "error", err)
	}
	if err == nil && links != nil {
		return links, nil
	}

	slog.Info("cache missed")
	links, err = s.r.ChatLinkRepo.GetLinksByChatID(ctx, id)
	if err != nil {
		slog.Error("get links by chat ID error", "error", err)
		return nil, fmt.Errorf("getting links by chat ID: %w", err)
	}

	err = s.cache.Set(ctx, id, links)
	if err != nil {
		slog.Error("setting links in cache error", "error", err)
	}

	return links, nil
}

func (s *ChatService) AddLink(ctx context.Context, chatID int64, req dto.AddLinkRequest) (*model.Link, error) {
	link := model.NewLink(req.Link, req.Tags)
	link, err := s.linkResolver.FormatLink(link)
	slog.Info("adding link service", "link", link)
	if err != nil {
		slog.Error("format link error", "error", err)
		return nil, fmt.Errorf("formatting link: %w", err)
	}

	var addedLink *model.Link

	err = s.TxManager.WithTx(ctx, func(ctx context.Context) error {
		addedLink, err = s.r.LinkRepo.AddLink(ctx, *link)
		if err != nil {
			slog.Error("error adding link", "error", err)
			return fmt.Errorf("adding link to repo: %w", err)
		}

		err = s.r.ChatLinkRepo.Subscribe(ctx, chatID, addedLink.ID)
		if err != nil {
			return fmt.Errorf("subscribing chat to link: %w", err)
		}

		for _, tag := range req.Tags {
			addedTag, saveTagErr := s.r.TagRepo.SaveTag(ctx, model.NewTag(tag))
			if saveTagErr != nil {
				return fmt.Errorf("error saving tag: %w", saveTagErr)
			}
			saveLinkTagErr := s.r.LinkTagRepo.Save(ctx, addedLink.ID, addedTag.ID)
			if saveLinkTagErr != nil {
				return fmt.Errorf("error adding tag to link: %w", saveLinkTagErr)
			}
		}
		addedLink.Tags = req.Tags

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("adding link transaction: %w", err)
	}

	err = s.cache.Invalidate(ctx, chatID)
	if err != nil {
		slog.Error("error invalidating cache", "error", err)
	}

	return addedLink, nil
}

func (s *ChatService) UpdateLink(ctx context.Context, link *model.Link) (*model.Link, error) {
	updatedLink, err := s.r.LinkRepo.UpdateLink(ctx, link.ID, link)
	if err != nil {
		return nil, fmt.Errorf("updating link: %w", err)
	}
	return updatedLink, nil
}

// TODO: лучше подчищать ссылки без подписчиков в фоновом режиме. но пока не усложняем
func (s *ChatService) Unsubscribe(ctx context.Context, chatID int64, req dto.DeleteLinkRequest) (*model.Link, error) {
	var link *model.Link
	var err error

	err = s.TxManager.WithTx(ctx, func(ctx context.Context) error {
		_, err = s.r.ChatRepo.GetChatByID(ctx, chatID)
		if err != nil {
			return fmt.Errorf("getting chat by ID: %w", err)
		}

		link, err = s.r.LinkRepo.GetLinkByName(ctx, strings.TrimSpace(req.Link))
		if err != nil {
			slog.Error("link not found in repo", "link", req.Link, "error", err)
			return model.ErrNotFound
		}

		err = s.r.ChatLinkRepo.Unsubscribe(ctx, chatID, link.ID)
		if err != nil {
			return fmt.Errorf("unsubscribing chat from link: %w", err)
		}

		subscribers, getSubscribersErr := s.r.ChatLinkRepo.GetChatIDsByLinkID(ctx, link.ID)
		if getSubscribersErr != nil {
			return fmt.Errorf("getting chats by link ID: %w", getSubscribersErr)
		}

		if len(subscribers) == 0 {
			deleteTagsErr := s.r.LinkTagRepo.DeleteTagsByLinkID(ctx, link.ID)
			if deleteTagsErr != nil {
				return fmt.Errorf("deleting tags by link ID: %w", deleteTagsErr)
			}
			deleted, deleteLinkErr := s.r.LinkRepo.DeleteLink(ctx, link.ID)
			if deleteLinkErr != nil {
				return fmt.Errorf("deleting link: %w", deleteLinkErr)
			}
			slog.Debug("deleted link", "link", deleted.Link)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unsubscribing link transaction: %w", err)
	}

	err = s.cache.Invalidate(ctx, chatID)
	if err != nil {
		slog.Error("error invalidating cache", "error", err)
	}

	return link, nil
}

func (s *ChatService) GetLinks(ctx context.Context) ([]model.Link, error) {
	links, err := s.r.LinkRepo.GetLinks(ctx, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("getting links: %w", err)
	}
	return links, nil
}

func (s *ChatService) GetLinksByChatIDAndTag(ctx context.Context, chatID int64, tag string) ([]model.Link, error) {
	var links []model.Link
	var err error

	if tag != "" {
		links, err = s.r.ChatLinkRepo.GetLinksByChatIDAndTag(ctx, chatID, tag)
		if err != nil {
			return nil, fmt.Errorf("getting links by chat ID and tag: %w", err)
		}
	} else {
		links, err = s.r.ChatLinkRepo.GetLinksByChatID(ctx, chatID)
		if err != nil {
			return nil, fmt.Errorf("getting links by chat ID: %w", err)
		}
	}

	for i := range links {
		var tags []model.Tag
		tags, err = s.r.LinkTagRepo.GetTagsByLinkID(ctx, links[i].ID)
		if err != nil {
			slog.Error("get tags by link id error", "link", links[i].Link, "error", err)
		}

		links[i].Tags = make([]string, 0)
		for _, currTag := range tags {
			links[i].Tags = append(links[i].Tags, currTag.Name)
		}
	}
	return links, nil
}
