package module

import "github.com/leg100/pug/internal/resource"

// logEnricher adds module related attributes to log records where pertinent.
type logEnricher struct {
	table *resource.Table[*Module]
}

// moduleResource is a resource that belongs to a module
type moduleResource interface {
	Module() *resource.Resource
}

// AddLogAttributes checks if one of the log args is a resource that
// belongs to a module, and if so, adds the module path to the log record.
func (e *logEnricher) AddLogAttributes(attrs ...any) []any {
	for _, arg := range attrs {
		res, ok := arg.(moduleResource)
		if !ok {
			continue
		}
		modResource := res.Module()
		if modResource == nil {
			continue
		}
		mod, err := e.table.Get(modResource.ID)
		if err != nil {
			continue
		}
		return append(attrs, "module", mod.Path)
	}
	return attrs
}
