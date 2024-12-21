package keys

import "github.com/charmbracelet/bubbles/key"

type common struct {
	Plan        key.Binding
	PlanDestroy key.Binding
	AutoApply   key.Binding
	Destroy     key.Binding
	Cancel      key.Binding
	Delete      key.Binding
	Execute     key.Binding
	State       key.Binding
	Retry       key.Binding
	Reload      key.Binding
	Edit        key.Binding
	Init        key.Binding
	InitUpgrade key.Binding
	Validate    key.Binding
	Format      key.Binding
	Cost        key.Binding
}

// Keys shared by several models.
var Common = common{
	Plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	PlanDestroy: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "plan destroy"),
	),
	AutoApply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "auto-apply"),
	),
	Destroy: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "destroy"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel"),
	),
	Delete: key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "delete"),
	),
	Execute: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "execute program"),
	),
	State: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "state"),
	),
	Retry: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "retry"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Edit: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "edit"),
	),
	Init: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "init"),
	),
	InitUpgrade: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "init -upgrade"),
	),
	Validate: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "validate"),
	),
	Format: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "format"),
	),
	Cost: key.NewBinding(
		key.WithKeys("$"),
		key.WithHelp("$", "cost"),
	),
}
