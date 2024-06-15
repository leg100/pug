package tui

import tea "github.com/charmbracelet/bubbletea"

func ReportError(err error) tea.Cmd {
	return CmdHandler(ErrorMsg(err))
}

type ErrorMsg error
