package split

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ToggleSplit   key.Binding
	IncreaseSplit key.Binding
	DecreaseSplit key.Binding
}

var Keys = keyMap{
	ToggleSplit: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "toggle split"),
	),
	DecreaseSplit: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "decrease split"),
	),
	IncreaseSplit: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "increase split"),
	),
}
