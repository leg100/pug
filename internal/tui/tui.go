package tui

import tea "github.com/charmbracelet/bubbletea"

type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (Model, tea.Cmd)
	Title() string
	View() string
}

func cmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type errorMsg struct {
	Error   error
	Message string
	Args    []any
}

func newErrorMsg(err error, msg string, args ...any) errorMsg {
	return errorMsg{
		Error:   err,
		Message: msg,
		Args:    args,
	}
}
