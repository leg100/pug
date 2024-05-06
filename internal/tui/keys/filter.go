package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type filter struct {
	Exit  key.Binding
	Clear key.Binding
}

// Filter is a key map of keys available in filter mode.
var Filter = filter{
	Exit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "exit filter"),
	),
	Clear: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),
}
