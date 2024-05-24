package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func ReportError(err error, msg string, args ...any) tea.Cmd {
	return CmdHandler(NewErrorMsg(err, msg, args...))
}

func ReportInfo(msg string, args ...any) tea.Cmd {
	return CmdHandler(InfoMsg(fmt.Sprintf(msg, args...)))
}

func OpenVim(path string) tea.Cmd {
	// TODO: use env var EDITOR
	// TODO: check for side effects of exec blocking the tui - do
	// messages get queued up?
	c := exec.Command("vim", path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return NewErrorMsg(err, "opening vim")
	})
}
