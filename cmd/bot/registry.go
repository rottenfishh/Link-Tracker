package bot

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
