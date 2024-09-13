package logging

import (
	"reflect"

	"github.com/leg100/pug/internal/resource"
)

// enricher enriches a log record with further meaningful attributes that aren't
// readily available to the caller.
type enricher struct {
	updaters []ArgsUpdater
}

func (e *enricher) AddArgsUpdater(updater ArgsUpdater) {
	e.updaters = append(e.updaters, updater)
}

func (e *enricher) enrich(args ...any) []any {
	for _, en := range e.updaters {
		args = en.UpdateArgs(args...)
	}
	return args
}

// ArgsUpdater updates a log message's arguments.
type ArgsUpdater interface {
	UpdateArgs(args ...any) []any
}

// ReferenceUpdater checks log arguments for references to T via its ID, either directly
// or via a struct field, and updates or adds T to the log arguments
// accordingly.
type ReferenceUpdater[T any] struct {
	Getter[T]

	Name  string
	Field string
}

type Getter[T any] interface {
	Get(resource.ID) (T, error)
}

func (e *ReferenceUpdater[T]) UpdateArgs(args ...any) []any {
	for i, arg := range args {
		// Where an argument is of type resource.ID, try and retrieve the
		// resource corresponding to the ID and replace argument with the
		// resource.
		if id, ok := arg.(resource.ID); ok {
			t, err := e.Get(id)
			if err != nil {
				continue
			}
			// replace id with T
			args[i] = t
			return args
		}
		// Where an argument is a struct (or a pointer to a struct), check if it
		// has a field matching the expect field name, with a corresponding
		// value of type resource.ID, and if so, try and retrieve resource with
		// that resource.ID and add it as a log argument preceded with e.Name.
		v := reflect.Indirect(reflect.ValueOf(arg))
		if v.Kind() != reflect.Struct {
			continue
		}
		f := reflect.Indirect(v.FieldByName(e.Field))
		if !f.IsValid() {
			continue
		}
		id, ok := f.Interface().(resource.ID)
		if !ok {
			continue
		}
		t, err := e.Get(id)
		if err != nil {
			// T with id does not exist
			continue
		}
		return append(args, e.Name, t)
	}
	return args
}
