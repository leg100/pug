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
	LeftPane         key.Binding
	TopRightPane     key.Binding
	BottomRightPane  key.Binding
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
		key.WithKeys("X"),
		key.WithHelp("X", "close pane"),
	),
	LeftPane: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "focus left pane"),
	),
	TopRightPane: key.NewBinding(
		key.WithKeys("ctrl+k"),
		key.WithHelp("ctrl+k", "focus top right pane"),
	),
	BottomRightPane: key.NewBinding(
		key.WithKeys("ctrl+j"),
		key.WithHelp("ctrl+j", "focus bottom right pane"),
	),
}
