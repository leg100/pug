package taskgroup

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	tuitask "github.com/leg100/pug/internal/tui/task"
)

// Maker makes taskgroup models
type Maker struct {
	TaskListMaker *tuitask.ListMaker
}

func (mm *Maker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	taskList, err := mm.TaskListMaker.Make(parent, width, height)
	if err != nil {
		return nil, err
	}

	m := model{Model: taskList}

	return m, nil
}

// model renders a taskgroup, listing its tasks.
type model struct {
	// Model is a task list model.
	tea.Model

	group   *task.Group
	helpers *tui.Helpers
}

func (m model) Title() string {
	return m.helpers.Breadcrumbs("TaskGroup", m.group)
}
