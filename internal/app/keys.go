package app

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Quit key.Binding
}

var Keys = keyMap{}

//func init() {
//	tui.RegisterHelpBindings(func(current tui.State) []key.Binding {
//		return []key.Binding{
//			Keys.Quit,
//		}
//	})
//}
