package task

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ToggleInfo key.Binding
	Enter      key.Binding
	ApplyPlan  key.Binding
	Raw        key.Binding
	Structured key.Binding
}

var localKeys = keyMap{
	ToggleInfo: key.NewBinding(
		key.WithKeys("I"),
		key.WithHelp("I", "toggle info"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view task"),
	),
	ApplyPlan: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply plan"),
	),
	Raw: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "raw"),
	),
	Structured: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "structured"),
	),
}

type groupListKeyMap struct {
	Enter key.Binding
}

var groupListKeys = groupListKeyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view group"),
	),
}
