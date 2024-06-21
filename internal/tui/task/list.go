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
	"github.com/leg100/pug/internal/tui/split"
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
		Width: task.MaxStatusLen,
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

// ListTaskMaker makes task models belonging to a task list model
type ListTaskMaker struct {
	*Maker
}

func (m *ListTaskMaker) Make(res resource.Resource, width, height int) (tea.Model, error) {
	return m.make(res, width, height, false)
}

// NewListMaker constructs a task list model maker
func NewListMaker(tasks tui.TaskService, runs tui.RunService, taskMaker *Maker, helpers *tui.Helpers) *ListMaker {
	return &ListMaker{
		TaskService: tasks,
		RunService:  runs,
		TaskMaker:   &ListTaskMaker{Maker: taskMaker},
		Helpers:     helpers,
	}
}

// ListMaker makes task list models
type ListMaker struct {
	RunService  tui.RunService
	TaskService tui.TaskService
	TaskMaker   tui.Maker
	Helpers     *tui.Helpers
}

func (mm *ListMaker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{
		table.ModuleColumn,
		table.WorkspaceColumn,
		commandColumn,
		statusColumn,
		runStatusColumn,
		runChangesColumn,
		ageColumn,
	}

	renderer := func(t *task.Task) table.RenderedRow {
		row := table.RenderedRow{
			table.ModuleColumn.Key:    mm.Helpers.ModulePath(t),
			table.WorkspaceColumn.Key: mm.Helpers.WorkspaceName(t),
			commandColumn.Key:         t.CommandString(),
			ageColumn.Key:             tui.Ago(time.Now(), t.Updated),
			table.IDColumn.Key:        t.String(),
			statusColumn.Key:          mm.Helpers.TaskStatus(t, false),
		}

		if rr := t.Run(); rr != nil {
			run := rr.(*runpkg.Run)
			if t.CommandString() == "plan" && run.PlanReport != nil {
				row[runChangesColumn.Key] = mm.Helpers.RunReport(*run.PlanReport, true)
			} else if t.CommandString() == "apply" && run.ApplyReport != nil {
				row[runChangesColumn.Key] = mm.Helpers.RunReport(*run.ApplyReport, true)
			}
			row[runStatusColumn.Key] = mm.Helpers.RunStatus(run, false)
		}

		return row
	}

	splitModel := split.New(split.Options[*task.Task]{
		Columns:      columns,
		Renderer:     renderer,
		TableOptions: []table.Option[*task.Task]{table.WithSortFunc(task.ByState)},
		Width:        width,
		Height:       height,
		Maker:        mm.TaskMaker,
	})
	m := List{
		Model:   splitModel,
		runs:    mm.RunService,
		tasks:   mm.TaskService,
		helpers: mm.Helpers,
	}
	return m, nil
}

type List struct {
	split.Model[*task.Task]

	runs    tui.RunService
	tasks   tui.TaskService
	helpers *tui.Helpers
}

func (m List) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.tasks.List(task.ListOptions{})
		return table.BulkInsertMsg[*task.Task](tasks)
	}
}

func (m List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Cancel):
			taskIDs := m.Table.SelectedOrCurrentIDs()
			return m, m.helpers.CreateTasks("cancel", m.tasks.Cancel, taskIDs...)
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.Table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskKind, tui.WithParent(row.Value))
			}
		case key.Matches(msg, keys.Common.Apply):
			runIDs, err := m.pruneApplyableTasks()
			if err != nil {
				return m, tui.ReportError(fmt.Errorf("applying tasks: %w", err))
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Apply %d plans?", len(runIDs)),
				m.helpers.CreateTasks("apply", m.runs.ApplyPlan, runIDs...),
			)
		case key.Matches(msg, keys.Common.State):
			if row, ok := m.Table.CurrentRow(); ok {
				if ws, ok := m.helpers.TaskWorkspace(row.Value); ok {
					return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws))
				} else {
					return m, tui.ReportError(errors.New("task not associated with a workspace"))
				}
			}
		case key.Matches(msg, keys.Common.Retry):
			taskIDs := m.Table.SelectedOrCurrentIDs()
			if len(taskIDs) == 0 {
				return m, nil
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Retry %d tasks?", len(taskIDs)),
				m.helpers.CreateTasks("retry", m.tasks.Retry, taskIDs...),
			)
		}
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m List) Title() string {
	return tui.Breadcrumbs("Tasks", resource.GlobalResource)
}

func (m List) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
		keys.Common.Apply,
		keys.Common.State,
		keys.Common.Retry,
	}
	return append(bindings, keys.KeyMapToSlice(split.Keys)...)
}

// pruneApplyableTasks removes from the selection any tasks that cannot be
// applied, i.e all tasks other than those that are a plan and are in the
// planned state. The run ID of each task after pruning is returned.
func (m *List) pruneApplyableTasks() ([]resource.ID, error) {
	return m.Table.Prune(func(task *task.Task) (resource.ID, bool) {
		rr := task.Run()
		if rr == nil {
			return resource.ID{}, true
		}
		run := rr.(*runpkg.Run)
		if run.Status != runpkg.Planned {
			return resource.ID{}, true
		}
		return run.ID, false
	})
}
