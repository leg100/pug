package workspace

import (
	"reflect"

	"github.com/leg100/pug/internal/resource"
)

// logEnricher adds module related attributes to log records where pertinent.
type logEnricher struct {
	table *resource.Table[*Workspace]
}

func (e *logEnricher) EnrichLogRecord(args ...any) []any {
	args = e.addWorkspaceName(args...)
	args = e.replaceIDWithWorkspace(args...)
	return args
}

// addWorkspacePath checks if one of the log message args is a struct with a WorkspaceID
// field, and if so, looks up the corresponding workspace and adds it to the
// message.
func (e *logEnricher) addWorkspaceName(args ...any) []any {
	for _, arg := range args {
		v := reflect.Indirect(reflect.ValueOf(arg))
		if v.Kind() != reflect.Struct {
			continue
		}
		f := v.FieldByName("WorkspaceID")
		if f.IsZero() {
			continue
		}
		id, ok := f.Interface().(resource.ID)
		if !ok {
			continue
		}
		ws, err := e.table.Get(id)
		if err != nil {
			// workspace with id does not exist
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
