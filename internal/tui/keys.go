package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ShrinkPaneHeight key.Binding
	GrowPaneHeight   key.Binding
	ShrinkPaneWidth  key.Binding
	GrowPaneWidth    key.Binding
	SwitchPane       key.Binding
	SwitchPaneBack   key.Binding
	ClosePane        key.Binding
	Explorer         key.Binding
	OutputPane       key.Binding
}

var Keys = keyMap{
	Explorer: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "focus explorer"),
	),
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
	SwitchPaneBack: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tabab", "switch last pane"),
	),
	ClosePane: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "close pane"),
	),
	OutputPane: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "focus output pane"),
	),
}
