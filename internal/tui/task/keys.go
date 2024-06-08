package task

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Info          key.Binding
	TogglePreview key.Binding
	GrowPreview   key.Binding
	ShrinkPreview key.Binding
	Follow        key.Binding
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
	GrowPreview: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "grow preview"),
	),
	ShrinkPreview: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "shrink preview"),
	),
	Follow: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "follow output"),
	),
}
