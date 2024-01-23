package internal

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit      key.Binding
	Help      key.Binding
	CloseHelp key.Binding
	Modules   key.Binding
	Init      key.Binding
	Plan      key.Binding
	Apply     key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("^c"),
		key.WithHelp("^c", "exit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	CloseHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	),
	Modules: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "modules"),
	),
	Init: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "init"),
	),
	Plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
}
