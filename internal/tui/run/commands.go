package run

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

// ApplyCommand creates a tea command for applying runs
func ApplyCommand(runs tui.RunService, runIDs ...resource.ID) tea.Cmd {
	switch len(runIDs) {
	case 0:
		return tui.ReportError(errors.New("no applyable runs found"), "")
	case 1:
		return tui.RequestConfirmation(
			"Proceed with apply",
			func() tea.Msg {
				run, err := runs.Get(runIDs[0])
				if err != nil {
					return tui.NewErrorMsg(err, "applying run")
				}
				if _, err := runs.Apply(runIDs[0]); err != nil {
					return tui.NewErrorMsg(err, "applying run")
				}
				// When one apply is triggered, the user is sent to the run page.
				return tui.NewNavigationMsg(tui.RunKind, tui.WithParent(run.Resource))
			},
		)
	default:
		return tui.RequestConfirmation(
			fmt.Sprintf("Apply %d runs", len(runIDs)),
			tui.CreateTasks("apply", runs.Apply, runIDs...),
		)
	}
}
