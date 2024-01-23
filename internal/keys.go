package internal

import "github.com/charmbracelet/bubbles/key"

var defaultKeys = []key.Binding{
	key.NewBinding(
		key.WithKeys("^c"),
		key.WithHelp("^c", "exit"),
	),
	key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "modules"),
	),
}
