package module

import (
	"reflect"

	"github.com/leg100/pug/internal/resource"
)

// logEnricher adds module related attributes to log records where pertinent.
type logEnricher struct {
	table *resource.Table[*Module]
}

func (e *logEnricher) EnrichLogRecord(args ...any) []any {
	args = e.addModulePath(args...)
	args = e.replaceIDWithModule(args...)
	return args
}

// addModulePath checks if one of the log message args is a struct with a ModuleID
// field, and if so, looks up the corresponding module and adds it to the
// message.
func (e *logEnricher) addModulePath(args ...any) []any {
	for _, arg := range args {
		v := reflect.Indirect(reflect.ValueOf(arg))
		if v.Kind() != reflect.Struct {
			continue
		}
		f := v.FieldByName("ModuleID")
		if !f.IsValid() {
			continue
		}
		id, ok := f.Interface().(resource.ID)
		if !ok {
			continue
		}
		mod, err := e.table.Get(id)
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
