package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ShrinkPaneHeight key.Binding
	GrowPaneHeight   key.Binding
	ShrinkPaneWidth  key.Binding
	GrowPaneWidth    key.Binding
	SwitchPane       key.Binding
	ClosePane        key.Binding
}

var Keys = keyMap{
	ShrinkPaneHeight: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "shrink pane height"),
	),
	GrowPaneHeight: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "grow pane height"),
	),
	ShrinkPaneWidth: key.NewBinding(
		key.WithKeys("<"),
		key.WithHelp("<", "shrink pane width"),
	),
	GrowPaneWidth: key.NewBinding(
		key.WithKeys(">"),
		key.WithHelp(">", "grow pane width"),
	),
	SwitchPane: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch pane"),
	),
	ClosePane: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close pane"),
	),
}
