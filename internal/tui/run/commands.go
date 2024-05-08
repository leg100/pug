package run

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
)

// CreateRuns creates a tea command for creating runs.
func CreateRuns(runs tui.RunService, opts run.CreateOptions, workspaceIDs ...resource.ID) tea.Cmd {
	if len(workspaceIDs) == 0 {
		return nil
	}
	return func() tea.Msg {
		var msg CreatedRunsMsg
		for _, wid := range workspaceIDs {
			run, err := runs.Create(wid, opts)
			if err != nil {
				msg.CreateErrs = append(msg.CreateErrs, err)
			}
			msg.Runs = append(msg.Runs, run)
		}
		return msg
	}
}

type CreatedRunsMsg struct {
	Runs []*run.Run
	// Errors from creating tasks
	CreateErrs []error
}

// ApplyCommand creates a tea command for applying runs
func ApplyCommand(runs tui.RunService, runIDs ...resource.ID) tea.Cmd {
	switch len(runIDs) {
	case 0:
		return tui.ReportError(errors.New("no applyable runs found"), "")
	case 1:
		return tui.YesNoPrompt("Proceed with apply?", func() tea.Msg {
			run, err := runs.Get(runIDs[0])
			if err != nil {
				return tui.NewErrorMsg(err, "applying run")
			}
			if _, err := runs.Apply(runIDs[0]); err != nil {
				return tui.NewErrorMsg(err, "applying run")
			}
			// When one apply is triggered, the user is sent to the run page.
			return tui.NewNavigationMsg(tui.RunKind, tui.WithParent(run.Resource))
		})
	default:
		return tui.YesNoPrompt(
			fmt.Sprintf("Apply %d runs (y/N)? ", len(runIDs)),
			tui.CreateTasks("apply", runs.Apply, runIDs...),
		)
	}
}
