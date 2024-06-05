package task

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

// GroupMaker makes taskgroup models
type GroupMaker struct {
	TaskService   tui.TaskService
	RunService    tui.RunService
	TaskListMaker *ListMaker
	Helpers       *tui.Helpers
}

func (mm *GroupMaker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	group, ok := parent.(*task.Group)
	if !ok {
		return nil, errors.New("expected taskgroup resource")
	}

	progress := progress.New(progress.WithDefaultGradient())
	progress.Width = 20

	m := groupModel{
		tasks: mm.TaskService,
		lp: newListPreview(listPreviewOptions{
			parent:      parent,
			width:       width,
			height:      height,
			runService:  mm.RunService,
			taskService: mm.TaskService,
			helpers:     mm.Helpers,
		}),
		progress: progress,
		group:    group,
		helpers:  mm.Helpers,
	}

	return m, nil
}

// groupModel is a model for a taskgroup, listing and previewing its tasks.
type groupModel struct {
	lp ListPreview

	// Progress bar showing how many tasks are complete
	progress progress.Model

	tasks   tui.TaskService
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
	case resource.Event[*task.Task]:
		// Only forward events for tasks that are part of this task group.
		if !slices.ContainsFunc(m.group.Tasks, func(t *task.Task) bool {
			return t.ID == msg.Payload.ID
		}) {
			return m, nil
		}
		// Update progress bar to reflect task status
		var finished int
		for _, t := range m.group.Tasks {
			if t.IsFinished() {
				finished++
			}
		}
		percentageComplete := float64(finished) / float64(len(m.group.Tasks))
		cmds = append(cmds, m.progress.SetPercent(percentageComplete))
	case progress.FrameMsg:
		// FrameMsg is sent when the progress bar wants to animate itself
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	// Forward message to wrapped list preview model
	m.lp, cmd = m.lp.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m groupModel) View() string {
	return m.lp.View()
}

func (m groupModel) Title() string {
	return m.helpers.Breadcrumbs("TaskGroup", m.group)
}

func (m groupModel) Status() string {
	return m.progress.View()
}
