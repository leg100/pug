package workspace

import "github.com/leg100/pug/internal/resource"

// logEnricher adds module related attributes to log records where pertinent.
type logEnricher struct {
	table *resource.Table[*Workspace]
}

// workspaceResource is a resource that belongs to a workspace
type workspaceResource interface {
	Workspace() *resource.Resource
}

// AddLogAttributes checks if one of the log args is a resource that belongs to
// a workspace, and if so, adds the workspace name to the log record.
func (e *logEnricher) AddLogAttributes(args ...any) []any {
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
		return append(args, "workspace", ws.Name)
	}
	return args
}
