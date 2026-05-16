package bot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	botservice "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/commands"
)

func TestDispatcher_StartAndHelpCommands_ReturnExpectedMessages(t *testing.T) {
	ctx := context.Background()
	scrapperClient := new(botservice.MockScrapperClient)
	scrapperClient.On("RegisterChat", mock.Anything, int64(777)).Return(nil).Once()

	dispatcher := botservice.NewDispatcher()
	dispatcher.Register(commands.NewStartCommand(scrapperClient))
	dispatcher.Register(&commands.HelpCommand{})
	dispatcher.Register(&commands.FallBackCommand{})

	startMessage, err := dispatcher.Dispatch(ctx, "/start", 777)
	require.NoError(t, err)
	require.Equal(t, "Добро пожаловать, путник. Введите /help для просмотра доступных команд.", startMessage.Text)

	helpMessage, err := dispatcher.Dispatch(ctx, "/help", 777)
	require.NoError(t, err)
	require.Equal(t, "Для начала работы, введите сначала /start\n"+
		"Доступные в боте команды:\n"+
		"/start - начать работу\n"+
		"/help - справка по боту\n"+
		"/track - отслеживать ссылку\n"+
		"/untrack <link> - перестать отслеживать ссылку\n"+
		"/list - посмотреть свои отслеживаемые ссылки\n"+
		"/cancel - отменить текущую команду\n", helpMessage.Text)

	scrapperClient.AssertExpectations(t)
}
