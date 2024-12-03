package task

import (
	"fmt"

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

// groupTaskMaker makes task models belonging to a task group model
type groupTaskMaker struct {
	*Maker
}

func (m *groupTaskMaker) Make(id resource.ID, width, height int) (tui.ChildModel, error) {
	return m.make(id, width, height, false)
}

// GroupMaker makes taskgroup models
type GroupMaker struct {
	taskListMaker *ListMaker
}

// NewGroupMaker constructs a task group model maker
func NewGroupMaker(tasks *task.Service, plans *plan.Service, taskMaker *Maker, helpers *tui.Helpers) *GroupMaker {
	return &GroupMaker{
		taskListMaker: &ListMaker{
			Tasks:     tasks,
			Plans:     plans,
			TaskMaker: &groupTaskMaker{Maker: taskMaker},
			Helpers:   helpers,
		},
	}
}

func (mm *GroupMaker) Make(id resource.ID, width, height int) (tui.ChildModel, error) {
	group, err := mm.taskListMaker.Tasks.GetGroup(id)
	if err != nil {
		return nil, err
	}

	list, err := mm.taskListMaker.Make(id, width, height)
	if err != nil {
		return nil, fmt.Errorf("making task list model: %w", err)
	}

	m := groupModel{
		ChildModel: list,
		group:      group,
		Helpers:    mm.taskListMaker.Helpers,
	}
	return &m, nil
}

// groupModel is a model for a taskgroup, listing and previewing its tasks.
type groupModel struct {
	tui.ChildModel
	*tui.Helpers

	group *task.Group
}

func (m groupModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	if len(m.group.CreateErrors) > 0 {
		err := fmt.Errorf("failed to create %d tasks: see logs", len(m.group.CreateErrors))
		cmds = append(cmds, tui.ReportError(err))
	}
	cmds = append(cmds, func() tea.Msg {
		// Seed table with task group's tasks
		return table.BulkInsertMsg[*task.Task](m.group.Tasks)
	})
	return tea.Batch(cmds...)
}

func (m *groupModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case table.BulkInsertMsg[*task.Task]:
		if m.skip(([]*task.Task)(msg)...) {
			return nil
		}
	case resource.Event[*task.Task]:
		if m.skip(msg.Payload) {
			return nil
		}
	}

	// Forward message to wrapped task list model
	cmds = append(cmds, m.ChildModel.Update(msg))

	return tea.Batch(cmds...)
}

// skip determines whether to skip forwarding the task to the wrapped task list
// model.
func (m *groupModel) skip(tasks ...*task.Task) bool {
	// If any of the tasks are not part of this task group then skip all tasks
	for _, t := range tasks {
		if !m.group.IncludesTask(t.ID) {
			return true
		}
	}
	return false
}

func (m groupModel) Title() string {
	return m.Breadcrumbs("TaskGroup", m.group)
}

func (m groupModel) Status() string {
	return m.GroupReport(m.group, false)
}

func (m groupModel) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
		keys.Common.Apply,
		keys.Common.State,
		keys.Common.Retry,
	}
	return append(bindings, keys.KeyMapToSlice(split.Keys)...)
}
