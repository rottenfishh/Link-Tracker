package in

import (
	"github.com/stretchr/testify/mock"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type MockTgAdapter struct {
	mock.Mock
	Updates chan model.ChatUpdate
}

func (a *MockTgAdapter) SendMessage(chatID int64, message *model.Message) error {
	a.Called(chatID, message)
	return nil
}

func (a *MockTgAdapter) GetUpdates() <-chan model.ChatUpdate {
	a.Called()
	return a.Updates
}
