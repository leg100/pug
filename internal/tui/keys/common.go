package keys

import "github.com/charmbracelet/bubbles/key"

type common struct {
	Plan      key.Binding
	Apply     key.Binding
	Cancel    key.Binding
	Delete    key.Binding
	ShowState key.Binding
	Retry     key.Binding
	Reload    key.Binding
	Module    key.Binding
	Workspace key.Binding
	Edit      key.Binding
	Init      key.Binding
	Validate  key.Binding
	Format    key.Binding
}

// Keys shared by several models.
var Common = common{
	Plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	ShowState: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "show state"),
	),
	Retry: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "retry"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Module: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "module"),
	),
	Workspace: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "workspace"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Init: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "init"),
	),
	Validate: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "validate"),
	),
	Format: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "format"),
	),
}
