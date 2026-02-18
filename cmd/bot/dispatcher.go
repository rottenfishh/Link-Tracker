package bot

import (
	"context"
	"fmt"
	"strings"
)

type Dispatcher struct {
	r *Registry
}

func NewDispatcher(r *Registry) *Dispatcher {
	return &Dispatcher{r: r}
}

func (d *Dispatcher) Dispatch(ctx *context.Context, command string) (*Message, error) {
	args := strings.Split(command, " ")
	if len(args) < 1 {
		return nil, fmt.Errorf("empty command")
	}
	cmd, _ := d.r.cmds[args[0]]
	msg, err := cmd.Execute(ctx, args[1:])
	if err != nil {
		return nil, err
	}
	return msg, nil
}
