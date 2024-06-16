package task

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ToggleInfo key.Binding
}

var localKeys = keyMap{
	ToggleInfo: key.NewBinding(
		key.WithKeys("I"),
		key.WithHelp("I", "toggle info"),
	),
}
