package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	runpkg "github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
)

var (
	commandColumn = table.Column{
		Key:        "command",
		Title:      "COMMAND",
		FlexFactor: 1,
	}
	statusColumn = table.Column{
		Key:   "task_status",
		Title: "STATUS",
		Width: runpkg.MaxStatusLen,
	}
	ageColumn = table.Column{
		Key:   "age",
		Title: "AGE",
		Width: 7,
	}
	runChangesColumn = table.Column{
		Key:        "run_changes",
		Title:      "RUN CHANGES",
		FlexFactor: 1,
	}
	runStatusColumn = table.Column{
		Key:   "run_status",
		Title: "RUN STATUS",
		Width: runpkg.MaxStatusLen,
	}
)

type ListMaker struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService
	TaskService      tui.TaskService
	MaxTasks         int
	Helpers          *tui.Helpers
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	var columns []table.Column
	// Add further columns depending upon the kind of parent
	switch parent.GetKind() {
	case resource.Global:
		// Show module and workspace columns in global tasks table
		columns = append(columns, table.ModuleColumn)
		fallthrough
	case resource.Module:
		// Show workspace column in module tasks table
		columns = append(columns, table.WorkspaceColumn)
	}
	columns = append(columns,
		commandColumn,
		statusColumn,
		runStatusColumn,
		runChangesColumn,
		ageColumn,
	)

	renderer := func(t *task.Task) table.RenderedRow {
		row := table.RenderedRow{
			table.ModuleColumn.Key:    m.Helpers.ModulePath(t),
			table.WorkspaceColumn.Key: m.Helpers.WorkspaceName(t),
			commandColumn.Key:         t.CommandString(),
			ageColumn.Key:             tui.Ago(time.Now(), t.Updated),
			table.IDColumn.Key:        t.String(),
			statusColumn.Key:          m.Helpers.TaskStatus(t),
		}

		if rr := t.Run(); rr != nil {
			run := rr.(*runpkg.Run)
			if t.CommandString() == "plan" && run.PlanReport != nil {
				row[runChangesColumn.Key] = m.Helpers.RunReport(*run.PlanReport)
			} else if t.CommandString() == "apply" && run.ApplyReport != nil {
				row[runChangesColumn.Key] = m.Helpers.RunReport(*run.ApplyReport)
			}
			row[runStatusColumn.Key] = m.Helpers.RunStatus(run)
		}

		return row
	}

	table := table.New(columns, renderer, width, height).
		WithSortFunc(task.ByState).
		WithParent(parent)

	return list{
		table:  table,
		svc:    m.TaskService,
		runs:   m.RunService,
		parent: parent,
		max:    m.MaxTasks,
	}, nil
}

type list struct {
	table  table.Model[resource.ID, *task.Task]
	svc    tui.TaskService
	runs   tui.RunService
	parent resource.Resource
	max    int
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.svc.List(task.ListOptions{
			Ancestor: m.parent.GetID(),
		})
		return table.BulkInsertMsg[*task.Task](tasks)
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskKind, tui.WithParent(row.Value))
			}
		case key.Matches(msg, keys.Common.Cancel):
			taskIDs := m.table.SelectedOrCurrentKeys()
			return m, CreateTasks("cancel", m.parent, m.svc.Cancel, taskIDs...)
		case key.Matches(msg, keys.Common.Apply):
			runIDs, err := m.pruneApplyableTasks()
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Apply %d plans?", len(runIDs)),
				CreateTasks("apply", m.parent, m.runs.ApplyPlan, runIDs...),
			)
		}
	}
	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// pruneApplyableTasks removes from the selection any tasks that cannot be
// applied, i.e all tasks other than those that are a plan and are in the
// planned state. The run ID of each task after pruning is returned.
func (m list) pruneApplyableTasks() ([]resource.ID, error) {
	runIDs, err := m.table.Prune(func(task *task.Task) (resource.ID, error) {
		rr := task.Run()
		if rr == nil {
			return resource.ID{}, errors.New("task is not applyable")
		}
		run := rr.(*runpkg.Run)
		if run.Status != runpkg.Planned {
			return resource.ID{}, errors.New("task run is not in the planned state")
		}
		return run.ID, nil
	})
	if err != nil {
		return nil, err
	}
	return runIDs, nil
}

func (m list) Title() string {
	return tui.GlobalBreadcrumb("Tasks", m.table.TotalString())
}

func (m list) View() string {
	return m.table.View()
}

func (m list) TabStatus() string {
	return fmt.Sprintf("(%s)", m.table.TotalString())
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Cancel,
	}
}
