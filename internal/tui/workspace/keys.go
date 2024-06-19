package workspace

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	SetCurrent key.Binding
}

var localKeys = keyMap{
	SetCurrent: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "set current"),
	),
}

type resourcesKeyMap struct {
	Taint   key.Binding
	Untaint key.Binding
	Move    key.Binding
	Reload  key.Binding
}

var resourcesKeys = resourcesKeyMap{
	Taint: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "taint"),
	),
	Untaint: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "untaint"),
	),
	Move: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "move"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
}
