package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	// MinHeight is the minimum height of the TUI.
	MinHeight = 24
	// FooterHeight is the height of the footer at the bottom of the TUI.
	FooterHeight = 1
	// MinContentHeight is the minimum height of content above the footer.
	MinContentHeight = MinHeight - FooterHeight
)

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
