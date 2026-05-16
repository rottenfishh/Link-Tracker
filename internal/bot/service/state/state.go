package state

type State struct {
	ChatID       int64
	StartCommand string
	Step         int
	Data         map[string]any
	UserArgs     []string
	EndCommand   string
}

func NewState(chatID int64, startCommand string) *State {
	return &State{ChatID: chatID, StartCommand: startCommand, Step: 1, Data: make(map[string]any)}
}
