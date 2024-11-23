package explorer

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Enter key.Binding
}

var localKeys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open/close folder"),
	),
}
