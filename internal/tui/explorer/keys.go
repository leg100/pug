package explorer

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Enter               key.Binding
	SetCurrentWorkspace key.Binding
	Execute             key.Binding
	ReloadModules       key.Binding
	ReloadWorkspaces    key.Binding
}

var localKeys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open/close folder"),
	),
	SetCurrentWorkspace: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "set current workspace"),
	),
	Execute: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "execute program"),
	),
	ReloadModules: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload modules"),
	),
	ReloadWorkspaces: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "reload workspaces"),
	),
}
