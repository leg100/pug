package tui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// NavigateTo sends an instruction to navigate to a page with the given model
// kind, and optionally parent resource.
func NavigateTo(kind Kind, opts ...NavigateOption) tea.Cmd {
	return CmdHandler(NewNavigationMsg(kind, opts...))
}

func ReportInfo(msg string, args ...any) tea.Cmd {
	return CmdHandler(InfoMsg(fmt.Sprintf(msg, args...)))
}

func OpenEditor(path string) tea.Cmd {
	// TODO: check for side effects of exec blocking the tui - do
	// messages get queued up?
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		return ReportError(errors.New("cannot open editor: environment variable EDITOR not set"))
	}
	cmd := exec.Command(editor, path)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return ReportError(fmt.Errorf("opening %s in editor: %w", path, err))()
		}
		return nil
	})
}
