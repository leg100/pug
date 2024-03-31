package keys

import "github.com/charmbracelet/bubbles/key"

type common struct {
	Plan      key.Binding
	Apply     key.Binding
	Cancel    key.Binding
	ShowState key.Binding
	Retry     key.Binding
	Reload    key.Binding
	Module    key.Binding
	Workspace key.Binding
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
}
