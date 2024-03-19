package tui

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Quit       key.Binding
	Edit       key.Binding
	Modules    key.Binding
	Workspaces key.Binding
	Runs       key.Binding
	Tasks      key.Binding
	Init       key.Binding
	Plan       key.Binding
	Apply      key.Binding
	Validate   key.Binding
	Format     key.Binding
	Cancel     key.Binding
	ShowState  key.Binding
	Retry      key.Binding
	Logs       key.Binding
	Escape     key.Binding
	Enter      key.Binding
	Help       key.Binding
	CloseHelp  key.Binding
	SelectAll  key.Binding
	Reload     key.Binding
	Tab        key.Binding
	TabLeft    key.Binding
}

var Keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("^c", "exit"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
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
	Validate: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "validate"),
	),
	Format: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "format"),
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
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab", "ctrl+pgdown"),
		key.WithHelp("tab", "forward tab"),
	),
	TabLeft: key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+pgup"),
		key.WithHelp("shift+tab", "back tab"),
	),
}

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

type generalKeyMap struct {
	Quit       key.Binding
	Modules    key.Binding
	Workspaces key.Binding
	Runs       key.Binding
	Tasks      key.Binding
	Logs       key.Binding
	Escape     key.Binding
	SelectAll  key.Binding
	Help       key.Binding
}

var GeneralKeys = generalKeyMap{
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
	Logs: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "logs"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc", "`"),
		key.WithHelp("esc, `", "back"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
