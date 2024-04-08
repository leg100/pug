package top

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Yes key.Binding
}

var localKeys = keyMap{
	Yes: key.NewBinding(
		key.WithKeys("y"),
	),
}
