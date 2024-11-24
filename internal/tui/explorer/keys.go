package explorer

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Enter               key.Binding
	SetCurrentWorkspace key.Binding
}

var localKeys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open/close folder"),
	),
	SetCurrentWorkspace: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "set current workspace"),
	),
}
