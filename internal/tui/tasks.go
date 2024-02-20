package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	taskpkg "github.com/leg100/pug/internal/task"
)

const tasksState State = "tasks"

type tasks struct {
	mod *module.Module

	list list.Model
}

func newTasks(mod *module.Module) tasks {
	// convert tasks to list items
	items := make([]list.Item, len(mod.Tasks))
	for i, t := range mod.Tasks {
		items[i] = t
	}

	delegate := newTaskDelegate(mod)
	l := list.New(items, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)

	return tasks{list: l, mod: mod}
}

func (m tasks) Init() tea.Cmd {
	return nil
}

func (m tasks) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Modules, Keys.Escape):
			return m, ChangeState(modulesState)
		}
	case ViewSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case taskNewMsg:
		task := (*taskpkg.Task)(msg)
		return m, m.list.InsertItem(0, task)
	}

	// Handle keyboard and mouse events in the viewport
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m tasks) Title() string {
	return fmt.Sprintf("tasks (%s)", m.mod)
}

func (m tasks) View() string {
	return m.list.View()
}
