package run

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	tuitask "github.com/leg100/pug/internal/tui/task"
)

// CreateRuns creates a tea command for creating runs and sending the user to
// the appropriate page.
func CreateRuns(runs tui.RunService, issuer resource.Resource, opts run.CreateOptions, workspaceIDs ...resource.ID) tea.Cmd {
	if len(workspaceIDs) == 0 {
		return nil
	}
	return func() tea.Msg {
		msg := CreatedRunsMsg{Issuer: issuer}
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
	// The parent resource of the page on which the request to create runs was
	// issued.
	Issuer resource.Resource
}

func HandleCreatedRuns(msg CreatedRunsMsg) (navigate tea.Cmd, info string, err error) {
	// Determine whether and where to navigate the user to.
	switch len(msg.Runs) {
	case 0:
		// No runs created, don't send user anywhere.
	case 1:
		// Send user directly to runs's page.
		navigate = tui.NavigateTo(tui.RunKind, tui.WithParent(msg.Runs[0]))
	default:
		// Multiple tasks. Send the user to the appropriate listing for the model kind that
		// issued the request to create tasks.
		var (
			opts = []tui.NavigateOption{tui.WithParent(msg.Issuer)}
			kind tui.Kind
		)
		switch msg.Issuer.GetKind() {
		case resource.Workspace:
			// Send user to the runs tab on the workspace page
			kind = tui.WorkspaceKind
			opts = append(opts, tui.WithTab(tui.RunsTabTitle))
		case resource.Module:
			// Send user to the runs tab on the module page
			kind = tui.ModuleKind
			opts = append(opts, tui.WithTab(tui.RunsTabTitle))
		default:
			// Send user to global runs listing
			kind = tui.RunListKind
		}
		navigate = tui.NavigateTo(kind, opts...)
	}

	if len(msg.Runs) == 1 {
		info = fmt.Sprintf("created %s successfully", msg.Runs[0])
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

// ApplyCommand creates a tea command for applying runs
func ApplyCommand(runs tui.RunService, issuer resource.Resource, runIDs ...resource.ID) tea.Cmd {
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
			return tui.NewNavigationMsg(tui.RunKind, tui.WithParent(run))
		})
	default:
		return tui.YesNoPrompt(
			fmt.Sprintf("Apply %d runs?", len(runIDs)),
			tuitask.CreateTasks(tuitask.ApplyCommand, issuer, runs.Apply, runIDs...),
		)
	}
}
