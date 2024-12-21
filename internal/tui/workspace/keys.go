package workspace

import (
	"github.com/charmbracelet/bubbles/key"
)

type resourcesKeyMap struct {
	Plan        key.Binding
	PlanDestroy key.Binding
	Apply       key.Binding
	Destroy     key.Binding
	Taint       key.Binding
	Untaint     key.Binding
	Move        key.Binding
	Reload      key.Binding
	Enter       key.Binding
}

var resourcesKeys = resourcesKeyMap{
	Plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "targeted plan"),
	),
	PlanDestroy: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "targeted plan destroy"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "targeted auto-apply"),
	),
	Destroy: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "targeted destroy"),
	),
	Taint: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "taint"),
	),
	Untaint: key.NewBinding(
		key.WithKeys("U"),
		key.WithHelp("U", "untaint"),
	),
	Move: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "move"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view resource"),
	),
}
