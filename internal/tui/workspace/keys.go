package workspace

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Edit       key.Binding
	Init       key.Binding
	Plan       key.Binding
	Apply      key.Binding
	Validate   key.Binding
	Format     key.Binding
	SetCurrent key.Binding
}

var localKeys = keyMap{
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Init: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "init"),
	),
	Plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	Validate: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "validate"),
	),
	Format: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "format"),
	),
	SetCurrent: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "set current"),
	),
}
