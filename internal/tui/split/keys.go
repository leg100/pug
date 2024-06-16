package split

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	TogglePreview key.Binding
	GrowPreview   key.Binding
	ShrinkPreview key.Binding
}

var Keys = keyMap{
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
}
