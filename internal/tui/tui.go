package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// 6 = header (3) + title (1) + border (1) + 1
const DefaultYPosition = 6

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
