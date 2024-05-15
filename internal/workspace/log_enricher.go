package workspace

import (
	"github.com/leg100/pug/internal/resource"
)

// logEnricher adds module related attributes to log records where pertinent.
type logEnricher struct {
	table *resource.Table[*Workspace]
}

// workspaceResource is a resource that belongs to a workspace
type workspaceResource interface {
	Workspace() *resource.Common
}

func (e *logEnricher) EnrichLogRecord(args ...any) []any {
	args = e.addWorkspaceName(args...)
	args = e.replaceIDWithWorkspace(args...)
	return args
}

// addWorkspaceName checks if one of the log args is a resource that belongs to
// a workspace, and if so, adds the workspace to the args
func (e *logEnricher) addWorkspaceName(args ...any) []any {
	for _, arg := range args {
		res, ok := arg.(workspaceResource)
		if !ok {
			continue
		}
		wsResource := res.Workspace()
		if wsResource == nil {
			continue
		}
		ws, err := e.table.Get(wsResource.ID)
		if err != nil {
			continue
		}
		return append(args, "workspace", ws)
	}
	return args
}

// replaceIDWithWorkspace checks if one of the arguments is a workspace ID, and
// if so, replaces it with a workspace instance.
func (e *logEnricher) replaceIDWithWorkspace(args ...any) []any {
	for i, arg := range args {
		id, ok := arg.(resource.ID)
		if !ok {
			// Not an id
			continue
		}
		ws, err := e.table.Get(id)
		if err != nil {
			// Not a workspace id
			continue
		}
		// replace id with workspace
		args[i] = ws
		return args
	}
	return args
}
