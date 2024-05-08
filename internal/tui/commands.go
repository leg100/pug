package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

// CreateTasks returns a command that creates one or more tasks using the given
// IDs. If a task fails to be created then no further tasks will be created, and
// an error notification is sent. If all tasks are successfully created then a status
// notification is sent accordingly.
func CreateTasks(cmd string, fn task.Func, ids ...resource.ID) tea.Cmd {
	if len(ids) == 0 {
		return nil
	}

	return func() tea.Msg {
		multi, errs := task.CreateMulti(fn, ids...)
		return CreatedTasksMsg{
			Command:    cmd,
			Tasks:      multi,
			CreateErrs: errs,
		}
	}
}

func WaitTasks(created CreatedTasksMsg) tea.Cmd {
	return func() tea.Msg {
		created.Tasks.Wait()
		return CompletedTasksMsg(created)
	}
}

// NavigateTo sends an instruction to navigate to a page with the given model
// kind, and optionally parent resource.
func NavigateTo(kind Kind, opts ...NavigateOption) tea.Cmd {
	return CmdHandler(NewNavigationMsg(kind, opts...))
}

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
