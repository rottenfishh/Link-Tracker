package in

import (
	"log"

	"github.com/stretchr/testify/mock"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type MockTgAdapter struct {
	mock.Mock
	Updates <-chan model.ChatUpdate
}

func (a *MockTgAdapter) SendMessage(chatID int64, message *model.Message) error {
	a.Called(chatID, message)
	return nil
}

func (a *MockTgAdapter) GetUpdates() <-chan model.ChatUpdate {
	args := a.Called()
	ch, ok := args.Get(0).(<-chan model.ChatUpdate)
	if !ok {
		log.Fatal("GetUpdates called with channel not of type model.ChatUpdate")
	}
	return ch
}
