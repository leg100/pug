package tui

import tea "github.com/charmbracelet/bubbletea"

type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (Model, tea.Cmd)
	Title() string
	View() string
}

func MsgHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type ErrorMsg struct {
	Error   error
	Message string
	Args    []any
}

func Err(err error, msg string, args ...any) ErrorMsg {
	return ErrorMsg{
		Error:   err,
		Message: msg,
		Args:    args,
	}
}
