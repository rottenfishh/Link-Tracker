package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Dispatcher struct {
	cmds      map[string]Command
	chatCache map[int64]*State
}

func NewDispatcher() *Dispatcher {
	d := &Dispatcher{cmds: make(map[string]Command), chatCache: make(map[int64]*State)}
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
func (r *Dispatcher) Dispatch(ctx context.Context, userMessage string, chatId int64) (*Message, error) {
	slog.Info(userMessage)

	userArgs := strings.Split(userMessage, " ")

	if len(userArgs) < 1 {
		return nil, fmt.Errorf("empty userMessage")
	}

	slog.Debug("Dispatching user's message", "userMessage: ", userMessage)

	var cmd Command

	state, err := r.getStateForChat(chatId, userArgs)
	if err != nil {
		cmd = r.cmds["/fallback"]
	} else {
		cmd = r.cmds[state.StartCommand]
	}

	answerMsg, err := cmd.Execute(ctx, state)
	if err != nil {
		slog.Error("Command executing", "error", err.Error())
		return nil, err
	}

	if answerMsg.IsFinished {
		delete(r.chatCache, chatId)
	}
	return answerMsg.Message, nil
}

func (r *Dispatcher) getStateForChat(chatId int64, userArgs []string) (*State, error) {
	if strings.HasPrefix(userArgs[0], "/") {
		_, ok := r.cmds[userArgs[0]]
		if !ok {
			return nil, fmt.Errorf("command not found: %s", userArgs[0])
		}

		slog.Debug("Starting new dialogue with user ", "chat id ", chatId)

		newState := NewState(chatId, userArgs[0])
		newState.UserArgs = userArgs[1:]
		r.chatCache[chatId] = newState

		return newState, nil
	} else {
		if state, ok := r.chatCache[chatId]; ok {
			slog.Debug("Continuing dialogue with user ", "chatId: ", chatId)
			state.UserArgs = userArgs
			return state, nil
		} else {
			return nil, fmt.Errorf("command not found: %s", userArgs[0])
		}
	}
}
