package commands

import "context"

type TrackCommand struct {
}

func (cmd *TrackCommand) Name() string {
	return "/track"
}

func (cmd *TrackCommand) Description() string {
	return "Command to start tracking events from a given link"
}

// TODO: check for no args
func (cmd *TrackCommand) Execute(ctx *context.Context, args []string) (*Message, error) {
	// call Scrapper.RegisterLink
	// check ifs not already tracked
	//После ввода /track бот получает ссылку на отслеживаемый ресурс.
	//    После получения ссылки бот запрашивает теги (необязательно). Теги -- это слова или короткие фразы, разделённые запятыми, по которым пользователь может классифицировать ресурс.
	//    Примеры: работа, баг, документация, рефакторинг.
	//    Пользователь может ввести /cancel, чтобы прервать процесс.
	//    Если пользователь отправляет другую команду — процесс отслеживания отменяется.
	return NewMessage("Succesfully started tracking events from link: " + args[0]), nil
}
