package internal

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newModuleDelegate(runner *runner) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)
	d.ShowDescription = false

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		mod, ok := m.SelectedItem().(module)
		if !ok {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.Init):
				return func() tea.Msg {
					task, err := runner.run(taskspec{
						prog:      "tofu",
						args:      []string{"init", "-input=false"},
						path:      mod.path,
						exclusive: isPluginCacheUsed(),
					})
					if err != nil {
						return taskFailedMsg(err.Error())
					}
					return newTaskMsg{
						mod:  mod,
						task: task,
					}
				}
			case key.Matches(msg, keys.Apply):
				return func() tea.Msg {
					task, err := runner.run(taskspec{
						prog: "tofu",
						args: []string{"apply", "-auto-approve", "-input=false"},
						path: mod.path,
					})
					if err != nil {
						return taskFailedMsg(err.Error())
					}
					return newTaskMsg{
						mod:  mod,
						task: task,
					}
				}
			}
		}

		return nil
	}

	return d
}
