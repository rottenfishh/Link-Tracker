package service

type State struct {
	ChatId       int64
	StartCommand string
	Step         int
	Data         map[string]any
	UserArgs     []string
	EndCommand   string
}

func NewState(chatId int64, startCommand string) *State {
	return &State{ChatId: chatId, StartCommand: startCommand, Step: 1, Data: make(map[string]any)}
}
