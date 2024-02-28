package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
)

const modulesState State = "modules"

func init() {
	registerHelpBindings(func(short bool, current State) []key.Binding {
		if current != modulesState {
			return []key.Binding{Keys.Modules}
		}
		return []key.Binding{
			Keys.Init,
			Keys.Plan,
			Keys.Apply,
			Keys.ShowState,
			Keys.Tasks,
		}
	})
}

type modules struct {
	list list.Model

	workdir string

	width  int
	height int
}

func newModules(svc *module.Service, workdir string) (modules, error) {
	if err := svc.Reload(); err != nil {
		return modules{}, err
	}
	mods := svc.List()
	items := make([]list.Item, len(mods))
	for i, mod := range mods {
		items[i] = mod
	}
	d := newDelegate(runner)
	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	return modules{list: l, workdir: workdir}, nil
}

func (m modules) Init() tea.Cmd {
	return nil
}

func (m modules) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GlobalKeyMsg:
		if msg.Current != modulesState {
			if key.Matches(msg.KeyMsg, Keys.Modules) {
				return m, ChangeState(modulesState)
			}
		}
	case ViewSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case taskFailedMsg:
		// TODO: update a status bar
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modules) Title() string {
	return fmt.Sprintf("modules (%s)", m.workdir)
}

func (m modules) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.list.View(),
	)
}

func newModuleDelegate(runner *taskpkg.Runner) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)
	d.ShowDescription = false

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		mod, ok := m.SelectedItem().(*module.Module)
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
