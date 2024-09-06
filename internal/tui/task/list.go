package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
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
)

// ListTaskMaker makes task models belonging to a task list model
type ListTaskMaker struct {
	*Maker
}

func (m *ListTaskMaker) Make(id resource.ID, width, height int) (tea.Model, error) {
	return m.make(id, width, height, false)
}

// NewListMaker constructs a task list model maker
func NewListMaker(tasks *task.Service, plans *plan.Service, taskMaker *Maker, helpers *tui.Helpers) *ListMaker {
	return &ListMaker{
		Tasks:     tasks,
		Plans:     plans,
		TaskMaker: &ListTaskMaker{Maker: taskMaker},
		Helpers:   helpers,
	}
}

// ListMaker makes task list models
type ListMaker struct {
	Plans *plan.Service
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
		table.SummaryColumn,
		ageColumn,
	}

	renderer := func(t *task.Task) table.RenderedRow {
		return table.RenderedRow{
			taskIDColumn.Key:          t.ID.String(),
			table.ModuleColumn.Key:    mm.Helpers.TaskModulePath(t),
			table.WorkspaceColumn.Key: mm.Helpers.TaskWorkspaceName(t),
			commandColumn.Key:         t.String(),
			ageColumn.Key:             tui.Ago(time.Now(), t.Updated),
			statusColumn.Key:          mm.Helpers.TaskStatus(t, false),
			table.SummaryColumn.Key:   mm.Helpers.TaskSummary(t, true),
		}
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
		plans:   mm.Plans,
		tasks:   mm.Tasks,
		Helpers: mm.Helpers,
	}
	return m, nil
}

type List struct {
	split.Model[*task.Task]
	*tui.Helpers

	plans *plan.Service
	tasks *task.Service
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
		case key.Matches(msg, localKeys.Enter):
			if row, ok := m.Table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskKind, tui.WithParent(row.ID))
			}
		case key.Matches(msg, keys.Common.Apply):
			specs, err := m.Table.Prune(func(t *task.Task) (task.Spec, error) {
				// Task must be a plan in order to be applied
				return m.plans.ApplyPlan(t.ID)
			})
			if err != nil {
				return m, tui.ReportError(fmt.Errorf("applying tasks: %w", err))
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Apply %d plans?", len(specs)),
				m.CreateTasksWithSpecs(specs...),
			)
		case key.Matches(msg, keys.Common.State):
			if row, ok := m.Table.CurrentRow(); ok {
				if ws := m.TaskWorkspaceOrCurrentWorkspace(row.Value); ws != nil {
					return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws.GetID()))
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
				m.CreateTasksWithSpecs(specs...),
			)
		}
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m List) Title() string {
	return m.Breadcrumbs("Tasks", resource.GlobalResource)
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
