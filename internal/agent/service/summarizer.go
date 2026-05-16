package service

import (
	"context"
	"log/slog"
)

type StupidSummarizer struct {
}

func (s *StupidSummarizer) Summarize(_ context.Context, description string, length int) (string, error) {
	slog.Info("stupid summarizer work")
	return description[:length], nil
}
