package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type global struct {
	Modules    key.Binding
	Workspaces key.Binding
	Runs       key.Binding
	Tasks      key.Binding
	Logs       key.Binding
	Escape     key.Binding
	Enter      key.Binding
	Select     key.Binding
	SelectAll  key.Binding
	Quit       key.Binding
	Help       key.Binding
}

var Global = global{
	Modules: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "all modules"),
	),
	Workspaces: key.NewBinding(
		key.WithKeys("W"),
		key.WithHelp("W", "all workspaces"),
	),
	Runs: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "all runs"),
	),
	Tasks: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "all tasks"),
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
		key.WithHelp("enter", "view"),
	),
	Select: key.NewBinding(
		key.WithKeys("select", " "),
		key.WithHelp("<space>", "select"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("^c", "exit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
