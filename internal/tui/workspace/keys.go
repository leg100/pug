package workspace

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	SetCurrent key.Binding
}

var localKeys = keyMap{
	SetCurrent: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "set current"),
	),
}
