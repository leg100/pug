package task

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

// groupTaskMaker makes task models belonging to a task group model
type groupTaskMaker struct {
	*Maker
}

func (m *groupTaskMaker) Make(res resource.Resource, width, height int) (tea.Model, error) {
	return m.makeWithID(res, width, height, TaskGroupMakerID, false)
}

// GroupMaker makes taskgroup models
type GroupMaker struct {
	TaskService tui.TaskService
	RunService  tui.RunService
	Helpers     *tui.Helpers

	taskListMaker *ListMaker
}

// NewGroupMaker constructs a task group model maker
func NewGroupMaker(tasks tui.TaskService, runs tui.RunService, taskMaker *Maker, helpers *tui.Helpers) *GroupMaker {
	return &GroupMaker{
		taskListMaker: &ListMaker{
			TaskService:       tasks,
			RunService:        runs,
			TaskMaker:         &groupTaskMaker{Maker: taskMaker},
			Helpers:           helpers,
			hideCommandColumn: true,
		},
	}
}

func (mm *GroupMaker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	group, ok := parent.(*task.Group)
	if !ok {
		return nil, errors.New("expected taskgroup resource")
	}

	progress := progress.New(progress.WithDefaultGradient())
	progress.Width = 20
	progress.ShowPercentage = false

	list, err := mm.taskListMaker.Make(parent, width, height)
	if err != nil {
		return nil, fmt.Errorf("making task list model: %w", err)
	}

	m := groupModel{
		Model:    list,
		progress: progress,
		group:    group,
		helpers:  mm.Helpers,
	}
	return m, nil
}

// groupModel is a model for a taskgroup, listing and previewing its tasks.
type groupModel struct {
	tea.Model

	// Progress bar showing how many tasks are complete
	progress progress.Model
	finished bool

	group   *task.Group
	helpers *tui.Helpers
}

func (m groupModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	if len(m.group.CreateErrors) > 0 {
		err := fmt.Errorf("failed to create %d tasks: see logs", len(m.group.CreateErrors))
		cmds = append(cmds, tui.ReportError(err, ""))
	}
	cmds = append(cmds, func() tea.Msg {
		// Seed table with task group's tasks
		return table.BulkInsertMsg[*task.Task](m.group.Tasks)
	})
	return tea.Batch(cmds...)
}

func (m groupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case table.BulkInsertMsg[*task.Task]:
		cmd, skip := m.handleTasks(([]*task.Task)(msg)...)
		if skip {
			return m, nil
		}
		cmds = append(cmds, cmd)
	case resource.Event[*task.Task]:
		cmd, skip := m.handleTasks(msg.Payload)
		if skip {
			return m, nil
		}
		cmds = append(cmds, cmd)
	case outputMsg:
		// Only forward output messages for tasks that are part of this task group.
		if !m.group.IncludesTask(msg.taskID) {
			return m, nil
		}
	case progress.FrameMsg:
		// FrameMsg is sent when the progress bar wants to animate itself
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	// Forward message to wrapped task list model
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m groupModel) Title() string {
	return m.helpers.Breadcrumbs("TaskGroup", m.group)
}

func (m groupModel) Status() string {
	pbar := m.progress.View()
	return fmt.Sprintf("%s %d/%d", pbar, m.group.Finished(), len(m.group.Tasks))
}

func (m *groupModel) handleTasks(tasks ...*task.Task) (tea.Cmd, bool) {
	if m.finished {
		return nil, true
	}

	for _, t := range tasks {
		// If any of the tasks are not part of this task group then skip
		// handling all tasks
		if !m.group.IncludesTask(t.ID) {
			return nil, true
		}
	}

	// Update progress bar to reflect task status
	percentageComplete := float64(m.group.Finished()) / float64(len(m.group.Tasks))
	if percentageComplete == 1 {
		// no more updates to progress bar necessary after this update.
		m.finished = true
	}
	return m.progress.SetPercent(percentageComplete), false
}
