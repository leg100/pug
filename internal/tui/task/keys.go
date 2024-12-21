package task

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ToggleInfo key.Binding
	Enter      key.Binding
	ApplyPlan  key.Binding
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
