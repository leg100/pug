package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/split"
	"github.com/leg100/pug/internal/tui/table"
)

var (
	taskIDColumn = table.Column{
		Key:   "task_id",
		Title: "TASK ID",
		Width: len("TASK_ID"),
	}
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
		Title:      "CHANGES",
		FlexFactor: 1,
	}
)

// ListTaskMaker makes task models belonging to a task list model
type ListTaskMaker struct {
	*Maker
}

func (m *ListTaskMaker) Make(id resource.ID, width, height int) (tea.Model, error) {
	return m.make(id, width, height, false)
}

// NewListMaker constructs a task list model maker
func NewListMaker(tasks *task.Service, runs *run.Service, taskMaker *Maker, helpers *tui.Helpers) *ListMaker {
	return &ListMaker{
		Tasks:     tasks,
		Runs:      runs,
		TaskMaker: &ListTaskMaker{Maker: taskMaker},
		Helpers:   helpers,
	}
}

// ListMaker makes task list models
type ListMaker struct {
	Runs  *run.Service
	Tasks *task.Service

	TaskMaker tui.Maker
	Helpers   *tui.Helpers
}

func (mm *ListMaker) Make(_ resource.ID, width, height int) (tea.Model, error) {
	columns := []table.Column{
		taskIDColumn,
		table.ModuleColumn,
		table.WorkspaceColumn,
		commandColumn,
		statusColumn,
		runChangesColumn,
		ageColumn,
	}

	renderer := func(t *task.Task) table.RenderedRow {
		row := table.RenderedRow{
			taskIDColumn.Key:          t.ID.String(),
			table.ModuleColumn.Key:    mm.Helpers.ModulePath(t),
			table.WorkspaceColumn.Key: mm.Helpers.WorkspaceName(t),
			commandColumn.Key:         t.String(),
			ageColumn.Key:             tui.Ago(time.Now(), t.Updated),
			statusColumn.Key:          mm.Helpers.TaskStatus(t, false),
		}

		if rr := t.Run(); rr != nil {
			run := rr.(*run.Run)
			if t.Command[0] == "plan" && run.PlanReport != nil {
				row[runChangesColumn.Key] = mm.Helpers.RunReport(*run.PlanReport, true)
			} else if t.Command[0] == "apply" && run.ApplyReport != nil {
				row[runChangesColumn.Key] = mm.Helpers.RunReport(*run.ApplyReport, true)
			}
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
		runs:    mm.Runs,
		tasks:   mm.Tasks,
		helpers: mm.Helpers,
	}
	return m, nil
}

type List struct {
	split.Model[*task.Task]

	runs    *run.Service
	tasks   *task.Service
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
			return m, cancel(m.tasks, taskIDs...)
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.Table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskKind, tui.WithParent(row.Value))
			}
		case key.Matches(msg, keys.Common.Apply):
			specs, err := m.Table.Prune(func(tsk *task.Task) (task.Spec, error) {
				// Task must belong to a run in order to be applied.
				res := tsk.Run()
				if res == nil {
					return task.Spec{}, errors.New("task does not belong to a run")
				}
				return m.runs.Apply(res.GetID(), nil)
			})
			if err != nil {
				return m, tui.ReportError(fmt.Errorf("applying tasks: %w", err))
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Apply %d plans?", len(specs)),
				m.helpers.CreateTasksWithSpecs(specs...),
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
			rows := m.Table.SelectedOrCurrent()
			specs := make([]task.Spec, len(rows))
			for i, row := range rows {
				specs[i] = row.Value.Spec
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Retry %d tasks?", len(rows)),
				m.helpers.CreateTasksWithSpecs(specs...),
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
