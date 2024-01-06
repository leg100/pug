package task

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Info key.Binding
}

var localKeys = keyMap{
	Info: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "info"),
	),
}
