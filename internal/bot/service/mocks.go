package service

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/mock"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

type MockScrapperClient struct {
	mock.Mock
}

func (m *MockScrapperClient) RegisterChat(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock register chat: %w", err)
	}
	return nil
}

func (m *MockScrapperClient) DeleteChat(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock delete chat: %w", err)
	}
	return nil
}

func (m *MockScrapperClient) RegisterLink(ctx context.Context, chatID int64, link string, tags []string) error {
	args := m.Called(ctx, chatID, link, tags)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock register link: %w", err)
	}
	return nil
}

func (m *MockScrapperClient) DeleteLink(ctx context.Context, chatID int64, link string) error {
	args := m.Called(ctx, chatID, link)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock delete link: %w", err)
	}
	return nil
}

func (m *MockScrapperClient) GetLinks(ctx context.Context, chatID int64, tag string) (*dto.ListLinksResponse, error) {
	args := m.Called(ctx, chatID, tag)

	var resp *dto.ListLinksResponse
	if v := args.Get(0); v != nil {
		resp, _ = v.(*dto.ListLinksResponse)
	}

	if err := args.Error(1); err != nil {
		return resp, fmt.Errorf("mock get links: %w", err)
	}
	return resp, nil
}
