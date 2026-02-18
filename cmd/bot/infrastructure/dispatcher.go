package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/cmd/bot/commands"
)

type Dispatcher struct {
	cmds map[string]commands.Command
}

func NewDispatcher() *Dispatcher {
	d := &Dispatcher{cmds: make(map[string]commands.Command)}
	return d
}

func (r *Dispatcher) Register(cmd commands.Command) {
	r.cmds[cmd.Name()] = cmd
}

func (r *Dispatcher) Delete(name string) {
	r.cmds[name] = nil
}

func (r *Dispatcher) Get(name string) commands.Command {
	return r.cmds[name]
}

func (r *Dispatcher) GetCommands() map[string]commands.Command {
	return r.cmds
}

// TODO: use tg-bot-api built-in parser of commands
func (r *Dispatcher) Dispatch(ctx *context.Context, command string) (*commands.Message, error) {
	args := strings.Split(command, " ")
	if len(args) < 1 {
		return nil, fmt.Errorf("empty command")
	}
	cmd, _ := r.cmds[args[0]]
	msg, err := cmd.Execute(ctx, args[1:])
	if err != nil {
		return nil, err
	}
	return msg, nil
}
