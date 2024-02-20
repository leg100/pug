package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	modulepkg "github.com/leg100/pug/internal/module"
	taskpkg "github.com/leg100/pug/internal/task"
)

func newDelegate(runner *taskpkg.Runner) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)
	d.ShowDescription = false

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		mod, ok := m.SelectedItem().(*modulepkg.Module)
		if !ok {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keys.Tasks, Keys.Enter):
				return ChangeState(tasksState, WithModelOption(
					newTasks(mod),
				))
			case key.Matches(msg, Keys.Init):
				return func() tea.Msg {
					t, err := mod.Init(runner)
					if err != nil {
						return Err(err, "creating init task", "module", mod)
					}
					return ChangeState(taskState, WithModelOption(
						newTask(t, mod, 0, 0),
					))()
				}
			case key.Matches(msg, Keys.Apply):
				return func() tea.Msg {
					t, err := mod.Apply(runner)
					if err != nil {
						return taskFailedMsg(err.Error())
					}
					return ChangeState(taskState, WithModelOption(
						newTask(t, mod, 0, 0),
					))()
				}
			case key.Matches(msg, Keys.ShowState):
				return func() tea.Msg {
					t, err := mod.ShowState(runner)
					if err != nil {
						return Err(err, "creating show state task", "module", mod)
					}
					return ChangeState(taskState, WithModelOption(
						newTask(t, mod, 0, 0),
					))()
				}
			}
		}

		return nil
	}

	return d
}
