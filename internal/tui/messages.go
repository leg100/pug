package tui

import tea "github.com/charmbracelet/bubbletea"

// navigationMsg is an instruction to navigate to a page.
type NavigationMsg struct {
	Target Page
}

func Navigate(target Page) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{Target: target}
	}
}

type ErrorMsg struct {
	Error   error
	Message string
	Args    []any
}

func NewErrorMsg(err error, msg string, args ...any) ErrorMsg {
	return ErrorMsg{
		Error:   err,
		Message: msg,
		Args:    args,
	}
}

func NewErrorCmd(err error, msg string, args ...any) tea.Cmd {
	return CmdHandler(NewErrorMsg(err, msg, args...))
}
