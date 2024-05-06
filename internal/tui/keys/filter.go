package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type filter struct {
	Blur  key.Binding
	Close key.Binding
}

// Filter is a key map of keys available in filter mode.
var Filter = filter{
	Blur: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "exit filter"),
	),
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),
}
