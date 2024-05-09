package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type global struct {
	Modules     key.Binding
	Workspaces  key.Binding
	Runs        key.Binding
	Tasks       key.Binding
	Logs        key.Binding
	Back        key.Binding
	Enter       key.Binding
	Select      key.Binding
	SelectAll   key.Binding
	SelectClear key.Binding
	SelectRange key.Binding
	Filter      key.Binding
	Quit        key.Binding
	Help        key.Binding
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
	Back: key.NewBinding(
		key.WithKeys("esc", "shift+tab"),
		key.WithHelp("esc/shift+tab", "back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view"),
	),
	Select: key.NewBinding(
		key.WithKeys("s", " "),
		key.WithHelp("<space>/s", "select"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
	SelectClear: key.NewBinding(
		key.WithKeys(`ctrl+\`),
		key.WithHelp(`ctrl+\`, "clear selection"),
	),
	SelectRange: key.NewBinding(
		key.WithKeys(`ctrl+@`),
		key.WithHelp(`ctrl+<space>`, "select range"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp(`/`, "filter"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
