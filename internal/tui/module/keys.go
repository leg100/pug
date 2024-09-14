package module

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	ReloadModules    key.Binding
	ReloadWorkspaces key.Binding
	Enter            key.Binding
	Execute          key.Binding
}

var localKeys = keyMap{
	ReloadModules: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload modules"),
	),
	ReloadWorkspaces: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "reload workspaces"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "state"),
	),
	Execute: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "execute program"),
	),
}
