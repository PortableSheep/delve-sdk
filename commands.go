package sdk

import "context"

// CommandMap is a convenience alias for registering multiple command handlers.
type CommandMap map[string]func(ctx context.Context, args []any) (any, error)

// RegisterCommands registers all command handlers in the provided map.
func (p *Plugin) RegisterCommands(cmds CommandMap) {
	for id, h := range cmds {
		p.OnCommand(id, h)
	}
}
