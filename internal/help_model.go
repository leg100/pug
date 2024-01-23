package internal

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type helpModel struct {
	last  model
	close bool
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keys.CloseHelp) {
			m.close = true
			return m, nil
		}
	}
	return m, nil
}

func (m helpModel) View() string {
	return ""
}

func (m helpModel) bindings() []key.Binding {
	return []key.Binding{keys.Init, keys.Plan, keys.Apply}
}
