package top

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
	tuirun "github.com/leg100/pug/internal/tui/run"
)

func handleCreatedRunsMsg(msg tuirun.CreatedRunsMsg) (navigate tea.Cmd, info string, err error) {
	if len(msg.Runs) == 1 {
		// User created one run successfully. Only in this scenario do we send
		// the user to the run page (in the scenario that *multiple* runs are
		// created then it is up to the original model to decide where to send
		// the user, whether to another page, to a tab, etc).
		info = fmt.Sprintf("created %s successfully", msg.Runs[0])
		navigate = tui.NavigateTo(tui.RunKind, tui.WithParent(msg.Runs[0].Resource))
	} else if len(msg.Runs) == 0 && len(msg.CreateErrs) == 1 {
		// User attempted to create one run but it failed to be created
		err = fmt.Errorf("creating run failed: %w", msg.CreateErrs[0])
	} else if len(msg.Runs) == 0 && len(msg.CreateErrs) > 1 {
		// User attempted to created multiple runs and they all failed to be
		// created
		err = fmt.Errorf("creating %d runs failed: see logs", len(msg.CreateErrs))
	} else if len(msg.CreateErrs) > 0 {
		// User attempted to create multiple runs and at least one failed to be
		// created, and at least one succeeded
		err = fmt.Errorf("created %d runs; %d failed to be created; see logs", len(msg.Runs), len(msg.CreateErrs))
	} else {
		// User attempted to create multiple runs and all were successfully
		// created.
		info = fmt.Sprintf("created %d runs successfully", len(msg.Runs))
	}
	return
}
