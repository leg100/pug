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
		table.ModuleColumn,
		table.WorkspaceColumn,
		commandColumn,
		statusColumn,
		table.SummaryColumn,
		ageColumn,
	}
	renderer := func(t *task.Task) table.RenderedRow {
		return table.RenderedRow{
			table.ModuleColumn.Key:    mm.Helpers.TaskModulePath(t),
			table.WorkspaceColumn.Key: mm.Helpers.TaskWorkspaceName(t),
			commandColumn.Key:         t.String(),
			ageColumn.Key:             tui.Ago(time.Now(), t.Created),
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
	m.common = &tui.ActionHandler{
		Helpers:     mm.Helpers,
		IDRetriever: &m,
	}
	return &m, nil
}

type List struct {
	table.Model[*task.Task]
	*tui.Helpers

	common *tui.ActionHandler
	plans  *plan.Service
	tasks  *task.Service
}

func (m *List) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.tasks.List(task.ListOptions{})
		return table.BulkInsertMsg[*task.Task](tasks)
	}
}

func (m *List) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Cancel):
			rows := m.SelectedOrCurrent()
			taskIDs := make([]resource.ID, len(rows))
			for i, row := range rows {
				taskIDs[i] = row.ID
			}
			return cancel(m.tasks, taskIDs...)
		case key.Matches(msg, keys.Common.AutoApply):
			ids, err := m.allPlans()
			if err != nil {
				return tui.ReportError(fmt.Errorf("applying tasks: %w", err))
			}
			return tui.YesNoPrompt(
				fmt.Sprintf("Apply %d plans?", len(ids)),
				m.CreateTasks(m.plans.ApplyPlan, ids...),
			)
		case key.Matches(msg, keys.Common.Retry):
			rows := m.SelectedOrCurrent()
			specs := make([]task.Spec, len(rows))
			for i, row := range rows {
				specs[i] = row.Spec
			}
			return tui.YesNoPrompt(
				fmt.Sprintf("Retry %d tasks?", len(rows)),
				m.CreateTasksWithSpecs(specs...),
			)
		default:
			cmd := m.common.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (m List) BorderText() map[tui.BorderPosition]string {
	return map[tui.BorderPosition]string{
		tui.TopLeftBorder:   tui.Bold.Render("tasks"),
		tui.TopMiddleBorder: m.Metadata(),
	}
}

func (m List) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
		keys.Common.State,
		keys.Common.Retry,
	}
	if _, err := m.allPlans(); err == nil {
		bindings = append(bindings, localKeys.ApplyPlan)
	}
	bindings = append(bindings, m.common.HelpBindings()...)
	return bindings
}

func (m List) allPlans() ([]resource.ID, error) {
	rows := m.SelectedOrCurrent()
	ids := make([]resource.ID, len(rows))
	for i, row := range rows {
		if err := plan.IsApplyable(row); err != nil {
			return nil, fmt.Errorf("at least one task is not applyable: %w", err)
		}
		ids[i] = row.ID
	}
	return ids, nil
}

func (m List) GetModuleIDs() ([]resource.ID, error) {
	rows := m.SelectedOrCurrent()
	ids := make([]resource.ID, len(rows))
	for i, row := range rows {
		if row.ModuleID == nil {
			return nil, errors.New("valid only on modules")
		}
		ids[i] = row.ModuleID
	}
	return ids, nil
}

func (m List) GetWorkspaceIDs() ([]resource.ID, error) {
	rows := m.SelectedOrCurrent()
	ids := make([]resource.ID, len(rows))
	for i, row := range rows {
		if row.WorkspaceID != nil {
			ids[i] = row.WorkspaceID
		} else if row.ModuleID == nil {
			return nil, errors.New("valid only on tasks associated with a module or a workspace")
		} else {
			// task has a module ID but no workspace ID, so find out if its
			// module has a current workspace, and if so, use that. Otherwise
			// return error
			mod, err := m.Modules.Get(row.ModuleID)
			if err != nil {
				return nil, err
			}
			if mod.CurrentWorkspaceID == nil {
				return nil, errors.New("valid only on tasks associated with a module with a current workspace, or a workspace")
			}
			ids[i] = mod.CurrentWorkspaceID
		}
	}
	return ids, nil
}
