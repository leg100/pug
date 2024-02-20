package tui

import tea "github.com/charmbracelet/bubbletea"

type (
	State string

	ChangeStateMsg struct {
		To    State
		Model Model
	}

	ChangeStateOption func(msg *ChangeStateMsg)
)

func ChangeState(to State, opts ...ChangeStateOption) tea.Cmd {
	return func() tea.Msg {
		msg := ChangeStateMsg{To: to}
		for _, f := range opts {
			f(&msg)
		}
		return msg
	}
}

func WithModelOption(m Model) ChangeStateOption {
	return func(msg *ChangeStateMsg) {
		msg.Model = m
	}
}
