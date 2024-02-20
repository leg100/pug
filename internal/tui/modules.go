package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	taskpkg "github.com/leg100/pug/internal/task"
)

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

const modulesState State = "modules"

type modules struct {
	list list.Model

	workdir string

	width  int
	height int
}

func newModules(runner *taskpkg.Runner) (modules, error) {
	// get mods and convert to items
	mods, err := module.FindModules(".")
	if err != nil {
		return modules{}, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return modules{}, err
	}
	items := make([]list.Item, len(mods))
	for i, mod := range mods {
		items[i] = mod
	}
	d := newDelegate(runner)
	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	return modules{list: l, workdir: wd}, nil
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
