package keys

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMapToSlice takes a struct of fields of type key.Binding and returns it as
// a slice instead.
func KeyMapToSlice(t any) (bindings []key.Binding) {
	typ := reflect.TypeOf(t)
	if typ.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < typ.NumField(); i++ {
		v := reflect.ValueOf(t).Field(i)
		bindings = append(bindings, v.Interface().(key.Binding))
	}
	return
}
