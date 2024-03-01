package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	taskpkg "github.com/leg100/pug/internal/task"
)

func newTaskDelegate(mod *module.Module) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)
	d.ShowDescription = true

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		task, ok := m.SelectedItem().(*taskpkg.Task)
		if !ok {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keys.Enter):
				return navigate(taskState, WithModelOption(
					newTaskModel(task, mod, 0, 0),
				))
			}
		}

		return nil
	}

	return d
}
