package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type navigation struct {
	LineUp       key.Binding
	LineDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	TabNext      key.Binding
	TabLast      key.Binding
	SwitchPane   key.Binding
}

// Navigation returns key bindings for navigation.
var Navigation = navigation{
	LineUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("b", "pgup"),
		key.WithHelp("b/pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("f", "pgdown"),
		key.WithHelp("f/pgdn", "page down"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("u", "ctrl+u"),
		key.WithHelp("u", "½ page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("d", "ctrl+d"),
		key.WithHelp("d", "½ page down"),
	),
	GotoTop: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
	),
	GotoBottom: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
	),
	TabNext: key.NewBinding(
		key.WithKeys("tab", "ctrl+pgdown"),
		key.WithHelp("tab/ctrl+pgdn", "next tab"),
	),
	TabLast: key.NewBinding(
		key.WithKeys("ctrl+pgup"),
		key.WithHelp("ctrl+pgup", "last tab"),
	),
	SwitchPane: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch pane"),
	),
}
