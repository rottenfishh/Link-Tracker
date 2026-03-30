package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/dto"
)

type MockScrapperClient struct {
	mock.Mock
}

func (m *MockScrapperClient) RegisterChat(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockScrapperClient) DeleteChat(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockScrapperClient) RegisterLink(ctx context.Context, chatID int64, link string, tags []string) error {
	args := m.Called(ctx, chatID, link, tags)
	return args.Error(0)
}

func (m *MockScrapperClient) DeleteLink(ctx context.Context, chatID int64, link string) error {
	args := m.Called(ctx, chatID, link)
	return args.Error(0)
}

func (m *MockScrapperClient) GetLinks(ctx context.Context, chatID int64, tag string) (*dto.ListLinksResponse, error) {
	args := m.Called(ctx, chatID, tag)

	var resp *dto.ListLinksResponse
	if v := args.Get(0); v != nil {
		resp = v.(*dto.ListLinksResponse)
	}

	return resp, args.Error(1)
}
