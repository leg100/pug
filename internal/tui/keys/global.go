package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type global struct {
	ShrinkPaneHeight key.Binding
	GrowPaneHeight   key.Binding
	ShrinkPaneWidth  key.Binding
	GrowPaneWidth    key.Binding
	ClosePane        key.Binding
	Explorer         key.Binding
	Tasks            key.Binding
	TaskGroups       key.Binding
	Logs             key.Binding
	Back             key.Binding
	Select           key.Binding
	SelectAll        key.Binding
	SelectClear      key.Binding
	SelectRange      key.Binding
	Filter           key.Binding
	Autoscroll       key.Binding
	Quit             key.Binding
	Suspend          key.Binding
	Help             key.Binding
}

var Global = global{
	ShrinkPaneHeight: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "shrink pane height"),
	),
	GrowPaneHeight: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "grow pane height"),
	),
	ShrinkPaneWidth: key.NewBinding(
		key.WithKeys("<"),
		key.WithHelp("<", "shrink pane width"),
	),
	GrowPaneWidth: key.NewBinding(
		key.WithKeys(">"),
		key.WithHelp(">", "grow pane width"),
	),
	ClosePane: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "close pane"),
	),
	Explorer: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "explorer"),
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
	Suspend: key.NewBinding(
		key.WithKeys("ctrl+z"),
		key.WithHelp("ctrl+z", "suspend"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	),
}
