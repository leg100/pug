package tui

import tea "github.com/charmbracelet/bubbletea"

type (
	State string

	ChangeStateOption func(msg *changeStateMsg)

	changeStateMsg struct {
		To    State
		Model Model
	}
)

func ChangeState(to State, opts ...ChangeStateOption) tea.Cmd {
	return func() tea.Msg {
		msg := changeStateMsg{To: to}
		for _, f := range opts {
			f(&msg)
		}
		return msg
	}
}

func WithModelOption(m Model) ChangeStateOption {
	return func(msg *changeStateMsg) {
		msg.Model = m
	}
}
