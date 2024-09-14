package keys

import "github.com/charmbracelet/bubbles/key"

type common struct {
	Plan        key.Binding
	PlanDestroy key.Binding
	Apply       key.Binding
	Destroy     key.Binding
	Cancel      key.Binding
	Delete      key.Binding
	State       key.Binding
	Retry       key.Binding
	Reload      key.Binding
	Module      key.Binding
	Workspace   key.Binding
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
		key.WithKeys("P"),
		key.WithHelp("P", "plan destroy"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	Destroy: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "destroy"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel"),
	),
	Delete: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "delete"),
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
