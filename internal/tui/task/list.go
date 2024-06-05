package task

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

type ListMaker struct {
	RunService  tui.RunService
	TaskService tui.TaskService
	Helpers     *tui.Helpers
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	list := List{
		lp: newListPreview(listPreviewOptions{
			parent:      parent,
			width:       width,
			height:      height,
			runService:  m.RunService,
			taskService: m.TaskService,
			helpers:     m.Helpers,
		}),
		taskService: m.TaskService,
		helpers:     m.Helpers,
	}
	return list, nil
}

type List struct {
	lp ListPreview

	taskService tui.TaskService
	helpers     *tui.Helpers
}

func (m List) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.taskService.List(task.ListOptions{})
		return table.BulkInsertMsg[*task.Task](tasks)
	}
}

func (m List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.lp, cmd = m.lp.Update(msg)
	return m, cmd
}

func (m List) View() string {
	return m.lp.View()
}

func (m List) Title() string {
	return tui.GlobalBreadcrumb("tasks", m.lp.list.TotalString())
}
