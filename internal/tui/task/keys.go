package task

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Info          key.Binding
	TogglePreview key.Binding
}

var localKeys = keyMap{
	Info: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "info"),
	),
	TogglePreview: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "toggle preview"),
	),
}
