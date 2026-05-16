package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/bot/service/state"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/pkg/model"
)

type Dispatcher struct {
	cmds      map[string]Command
	chatCache map[int64]*state.State
}

func NewDispatcher() *Dispatcher {
	d := &Dispatcher{cmds: make(map[string]Command), chatCache: make(map[int64]*state.State)}
	return d
}

func (r *Dispatcher) Register(cmd Command) {
	r.cmds[cmd.Name()] = cmd
}

func (r *Dispatcher) Delete(name string) {
	r.cmds[name] = nil
}

func (r *Dispatcher) Get(name string) Command {
	return r.cmds[name]
}

func (r *Dispatcher) GetCommands() map[string]Command {
	return r.cmds
}

// TODO: use tg-bot-docs built-in parser of commands
func (r *Dispatcher) Dispatch(ctx context.Context, userMessage string, chatID int64) (*model.Message, error) {
	slog.Info(userMessage)

	userArgs := strings.Split(userMessage, " ")

	if len(userArgs) < 1 {
		return nil, errors.New("empty userMessage")
	}

	slog.Debug("Dispatching user's message", "userMessage: ", userMessage)

	var cmd Command

	currState, err := r.getStateForChat(chatID, userArgs)
	if err != nil {
		cmd = r.cmds["/fallback"]
	} else {
		cmd = r.cmds[currState.StartCommand]
	}

	answerMsg, err := cmd.Execute(ctx, currState)
	if err != nil {
		slog.Error("Command executing", "error", err.Error())
		return nil, fmt.Errorf("executing command %s: %w", cmd.Name(), err)
	}

	if answerMsg.IsFinished() {
		delete(r.chatCache, chatID)
	}
	return answerMsg.Message, nil
}

func (r *Dispatcher) getStateForChat(chatID int64, userArgs []string) (*state.State, error) {
	if strings.HasPrefix(userArgs[0], "/") {
		_, ok := r.cmds[userArgs[0]]
		if !ok {
			return nil, fmt.Errorf("command not found: %s", userArgs[0])
		}

		slog.Debug("Starting new dialogue with user ", "chat id ", chatID)

		newState := state.NewState(chatID, userArgs[0])
		newState.UserArgs = userArgs[1:]
		r.chatCache[chatID] = newState

		return newState, nil
	}

	if currState, ok := r.chatCache[chatID]; ok {
		slog.Debug("Continuing dialogue with user ", "chatId: ", chatID)
		currState.UserArgs = userArgs
		return currState, nil
	}

	return nil, fmt.Errorf("command not found: %s", userArgs[0])
}
