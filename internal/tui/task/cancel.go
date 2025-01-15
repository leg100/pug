package task

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
)

// cancel task(s)
func cancel(tasks *task.Service, taskIDs ...resource.Identity) tea.Cmd {
	var (
		prompt string
		cmd    tea.Cmd
	)
	switch len(taskIDs) {
	case 0:
		return nil
	case 1:
		prompt = "Cancel task?"
		cmd = func() tea.Msg {
			if _, err := tasks.Cancel(taskIDs[0]); err != nil {
				return tui.ErrorMsg(fmt.Errorf("cancelling task: %w", err))
			}
			return tui.InfoMsg("sent cancel signal to task")
		}
	default:
		prompt = fmt.Sprintf("Cancel %d tasks?", len(taskIDs))
		cmd = func() tea.Msg {
			var errored bool
			for _, id := range taskIDs {
				if _, err := tasks.Cancel(id); err != nil {
					errored = true
				}
			}
			if errored {
				return tui.ErrorMsg(errors.New("one or more cancel requests failed; see logs"))
			}
			return tui.InfoMsg(fmt.Sprintf("sent cancel signal to %d tasks", len(taskIDs)))
		}
	}
	return tui.YesNoPrompt(prompt, cmd)
}
