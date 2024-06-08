package task

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Info          key.Binding
	TogglePreview key.Binding
	GrowPreview   key.Binding
	ShrinkPreview key.Binding
	Autoscroll    key.Binding
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
	Autoscroll: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "toggle autoscroll"),
	),
}
