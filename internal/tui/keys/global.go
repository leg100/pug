package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type global struct {
	Modules     key.Binding
	Workspaces  key.Binding
	Tasks       key.Binding
	TaskGroups  key.Binding
	Logs        key.Binding
	Back        key.Binding
	Select      key.Binding
	SelectAll   key.Binding
	SelectClear key.Binding
	SelectRange key.Binding
	Filter      key.Binding
	Autoscroll  key.Binding
	Quit        key.Binding
	Help        key.Binding
}

var Global = global{
	Modules: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "modules"),
	),
	Workspaces: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "workspaces"),
	),
	Tasks: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "tasks"),
	),
	TaskGroups: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "taskgroups"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "logs"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("<space>", "select"),
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
	Autoscroll: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "toggle autoscroll"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	),
}
