package bot

import (
	"context"
	"fmt"
	"strings"
)

type Dispatcher struct {
	cmds map[string]Command
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{cmds: make(map[string]Command)}
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

func (r *Dispatcher) Dispatch(ctx *context.Context, command string) (*Message, error) {
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
