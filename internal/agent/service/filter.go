package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Producer interface {
	SendUpdate(ctx context.Context, update model.LinkUpdate) error
}

type Summarizer interface {
	Summarize(ctx context.Context, description string, length int) (string, error)
}

type FilterConfig struct {
	StopWords     []string `config:"stop-words"`
	BannedAuthors []string `config:"excluded-authors"`
	MinLength     int      `config:"min-length"`
	Summarization struct {
		Threshold int `config:"threshold"`
	} `config:"summarization"`
}

type Filter struct {
	cfg        FilterConfig
	summarizer Summarizer
	producer   Producer
}

func NewFilter(cfg FilterConfig, summarizer Summarizer, producer Producer) *Filter {
	return &Filter{cfg: cfg, summarizer: summarizer, producer: producer}
}

func (f *Filter) HandleUpdate(ctx context.Context, upd model.LinkUpdate) error {
	ok := f.reviewUpdate(upd)
	if !ok {
		slog.Info("filtered message", "message", upd)
		return nil
	}

	slog.Info("len", len(upd.Update.Description), "threshold", f.cfg.Summarization.Threshold)
	if len(upd.Update.Description) > f.cfg.Summarization.Threshold {
		desc, err := f.summarizer.Summarize(ctx, upd.Update.Description, f.cfg.Summarization.Threshold)
		slog.Info("too long message")
		if err != nil {
			slog.Error("summarize error", "error", err)
		}
		upd.Update.Description = desc
	}

	err := f.producer.SendUpdate(ctx, upd)
	if err != nil {
		return fmt.Errorf("sending update: %w", err)
	}

	return nil
}

func (f *Filter) reviewUpdate(upd model.LinkUpdate) bool {
	if len(upd.Update.Description) < f.cfg.MinLength {
		slog.Info("not enough length")
		return false
	}

	if slices.Contains(f.cfg.BannedAuthors, upd.Update.Author) {
		slog.Info("banned author", "author", upd.Update.Author)
		return false
	}

	upd.Update.Description = strings.ToLower(upd.Update.Description)
	for _, banWord := range f.cfg.StopWords {
		if strings.Contains(upd.Update.Description, strings.ToLower(banWord)) {
			slog.Info("banned word", "word", banWord)
			return false
		}
	}

	return true
}
