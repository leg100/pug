package workspace

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Edit     key.Binding
	Init     key.Binding
	Plan     key.Binding
	Apply    key.Binding
	Validate key.Binding
	Format   key.Binding
	Reload   key.Binding
	Delete   key.Binding
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
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
}
