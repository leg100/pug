package common

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Quit       key.Binding
	Modules    key.Binding
	Workspaces key.Binding
	Runs       key.Binding
	Tasks      key.Binding
	Init       key.Binding
	Plan       key.Binding
	Apply      key.Binding
	Cancel     key.Binding
	ShowState  key.Binding
	Retry      key.Binding
	Logs       key.Binding
	Escape     key.Binding
	Enter      key.Binding
	Help       key.Binding
	CloseHelp  key.Binding
}

var Keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("^c", "exit"),
	),
	Modules: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "modules"),
	),
	Workspaces: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "workspaces"),
	),
	Runs: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "runs"),
	),
	Tasks: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "tasks"),
	),
	Init: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "init"),
	),
	Plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel"),
	),
	ShowState: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "show state"),
	),
	Retry: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "retry"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "logs"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc", "`"),
		key.WithHelp("esc, `", "back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	CloseHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	),
}

// keyMapToSlice takes a struct of fields of type key.Binding and returns it as
// a slice instead.
func keyMapToSlice(t any) (bindings []key.Binding) {
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
