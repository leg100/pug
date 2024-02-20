package help

import (
	"github.com/charmbracelet/bubbles/key"
)

var (
	Key = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	)
	CloseKey = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	)
)
