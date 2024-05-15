package module

import (
	"github.com/leg100/pug/internal/resource"
)

// logEnricher adds module related attributes to log records where pertinent.
type logEnricher struct {
	table *resource.Table[*Module]
}

// moduleResource is a resource that belongs to a module
type moduleResource interface {
	Module() *resource.Common
}

func (e *logEnricher) EnrichLogRecord(args ...any) []any {
	args = e.addModulePath(args...)
	args = e.replaceIDWithModule(args...)
	return args
}

// addModulePath checks if one of the log args is a resource that
// belongs to a module, and if so, adds the module to the args
func (e *logEnricher) addModulePath(args ...any) []any {
	for _, arg := range args {
		res, ok := arg.(moduleResource)
		if !ok {
			// does not belong to a module
			continue
		}
		modResource := res.Module()
		if modResource == nil {
			// can belong to a module but not in this instance
			continue
		}
		mod, err := e.table.Get(modResource.ID)
		if err != nil {
			// module with id does not exist
			continue
		}
		return append(args, "module", mod)
	}
	return args
}

// replaceIDWithModule checks if one of the arguments is a module ID, and
// if so, replaces it with a module instance.
func (e *logEnricher) replaceIDWithModule(args ...any) []any {
	for i, arg := range args {
		id, ok := arg.(resource.ID)
		if !ok {
			// Not an id
			continue
		}
		mod, err := e.table.Get(id)
		if err != nil {
			// Not a module id
			continue
		}
		// replace id with module
		args[i] = mod
		return args
	}
	return args
}
