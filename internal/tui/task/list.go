package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func (m *ListTaskMaker) Make(id resource.ID, width, height int) (tui.ChildModel, error) {
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

func (mm *ListMaker) Make(_ resource.ID, width, height int) (tui.ChildModel, error) {
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
			statusColumn.Key:          mm.Helpers.TaskStatus(t, true),
			table.SummaryColumn.Key:   mm.Helpers.TaskSummary(t, true),
		}
	}
	tbl := table.New(
		columns,
		renderer,
		width,
		height,
		table.WithSortFunc(task.ByState),
		table.WithPreview[*task.Task](tui.TaskKind),
	)
	m := List{
		Model:   tbl,
		plans:   mm.Plans,
		tasks:   mm.Tasks,
		Helpers: mm.Helpers,
	}
	return &m, nil
}

type List struct {
	table.Model[*task.Task]
	*tui.Helpers

	plans *plan.Service
	tasks *task.Service
}

func (m *List) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.tasks.List(task.ListOptions{})
		return table.BulkInsertMsg[*task.Task](tasks)
	}
}

func (m *List) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Cancel):
			taskIDs := m.SelectedOrCurrentIDs()
			return cancel(m.tasks, taskIDs...)
		case key.Matches(msg, localKeys.Enter):
			if row, ok := m.CurrentRow(); ok {
				return tui.NavigateTo(tui.TaskKind, tui.WithParent(row.ID), tui.WithPosition(tui.BottomRightPane))
			}
		case key.Matches(msg, keys.Common.Apply):
			specs, err := m.Prune(func(t *task.Task) (task.Spec, error) {
				// Task must be a plan in order to be applied
				return m.plans.ApplyPlan(t.ID)
			})
			if err != nil {
				return tui.ReportError(fmt.Errorf("applying tasks: %w", err))
			}
			return tui.YesNoPrompt(
				fmt.Sprintf("Apply %d plans?", len(specs)),
				m.CreateTasksWithSpecs(specs...),
			)
		case key.Matches(msg, keys.Common.State):
			if row, ok := m.CurrentRow(); ok {
				if ws := m.TaskWorkspaceOrCurrentWorkspace(row.Value); ws != nil {
					return tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws.GetID()))
				} else {
					return tui.ReportError(errors.New("task not associated with a workspace"))
				}
			}
		case key.Matches(msg, keys.Common.Retry):
			rows := m.SelectedOrCurrent()
			specs := make([]task.Spec, len(rows))
			for i, row := range rows {
				specs[i] = row.Value.Spec
			}
			return tui.YesNoPrompt(
				fmt.Sprintf("Retry %d tasks?", len(rows)),
				m.CreateTasksWithSpecs(specs...),
			)
		}
	}
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return cmd
}

func (m List) BorderText() map[tui.BorderPosition]string {
	t := lipgloss.NewStyle().
		Foreground(tui.DarkRed).
		Render("t")
	return map[tui.BorderPosition]string{
		tui.TopLeft:   fmt.Sprintf("[%sasks]", t),
		tui.TopMiddle: m.Metadata(),
	}
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
