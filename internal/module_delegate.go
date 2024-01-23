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
	keys := newDelegateKeyMap()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		mod, ok := m.SelectedItem().(module)
		if !ok {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.init):
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
			case key.Matches(msg, keys.apply):
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

	help := []key.Binding{keys.init, keys.apply}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	init  key.Binding
	apply key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.init,
		d.apply,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{d.init},
		{d.apply},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		init: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "init"),
		),
		apply: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "apply"),
		),
	}
}
