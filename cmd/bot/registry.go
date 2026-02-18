package bot

import (
	"context"
	"fmt"
	"strings"
)

type Registry struct {
	cmds map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{cmds: make(map[string]Command)}
}

func (r *Registry) Register(cmd Command) {
	r.cmds[cmd.Name()] = cmd
}

func (r *Registry) Delete(name string) {
	r.cmds[name] = nil
}

func (r *Registry) Get(name string) Command {
	return r.cmds[name]
}

func (r *Registry) GetCommands() map[string]Command {
	return r.cmds
}

func (r *Registry) Dispatch(ctx *context.Context, command string) (*Message, error) {
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
