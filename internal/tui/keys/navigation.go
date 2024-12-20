package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type navigation struct {
	LineUp          key.Binding
	LineDown        key.Binding
	PageUp          key.Binding
	PageDown        key.Binding
	HalfPageUp      key.Binding
	HalfPageDown    key.Binding
	GotoTop         key.Binding
	GotoBottom      key.Binding
	SwitchPane      key.Binding
	SwitchPaneBack  key.Binding
	LeftPane        key.Binding
	TopRightPane    key.Binding
	BottomRightPane key.Binding
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
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdn", "page down"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "½ page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "½ page down"),
	),
	GotoTop: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
	),
	GotoBottom: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
	),
	SwitchPane: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	),
	SwitchPaneBack: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "last pane"),
	),
	LeftPane: key.NewBinding(
		key.WithKeys("0"),
		key.WithHelp("0", "left pane"),
	),
	TopRightPane: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "top right pane"),
	),
	BottomRightPane: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "bottom right pane"),
	),
}
