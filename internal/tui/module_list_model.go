package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/workspace"
)

const moduleListState State = "modules"

func init() {
	registerHelpBindings(func(short bool, current State) []key.Binding {
		if current != moduleListState {
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

type moduleListModel struct {
	list list.Model

	workdir string

	width  int
	height int
}

func newModuleListModel(svc *module.Service, workdir string) (moduleListModel, error) {
	if err := svc.Reload(); err != nil {
		return moduleListModel{}, err
	}
	mods := svc.List()
	items := make([]list.Item, len(mods))
	for i, mod := range mods {
		items[i] = mod
	}
	d := newModuleDelegate(runner)
	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	return moduleListModel{list: l, workdir: workdir}, nil
}

func (moduleListModel) Init() tea.Cmd {
	return nil
}

func (mlm moduleListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case globalKeyMsg:
		if msg.Current != moduleListState {
			if key.Matches(msg.KeyMsg, Keys.Modules) {
				return mlm, ChangeState(moduleListState)
			}
		}
	case viewSizeMsg:
		mlm.list.SetSize(msg.Width, msg.Height)
		mlm.width = msg.Width
		mlm.height = msg.Height
		return mlm, nil
	case taskFailedMsg:
		// TODO: update a status bar
		return mlm, tea.Quit
	}

	var cmd tea.Cmd
	mlm.list, cmd = mlm.list.Update(msg)
	return mlm, cmd
}

func (mlm moduleListModel) Title() string {
	return fmt.Sprintf("modules (%s)", mlm.workdir)
}

func (mlm moduleListModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		mlm.list.View(),
	)
}

func newModuleDelegate(svc *module.Service, runs *run.Service, workspaces *workspace.Service) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)
	d.ShowDescription = false

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		mod, ok := m.SelectedItem().(*module.Module)
		if !ok {
			return nil
		}
		currentWorkspace, err := workspaces.GetCurrent(mod.ID)
		if err != nil {
			return cmdHandler(newErrorMsg(err, "creating init task", "module", mod))
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keys.Tasks, Keys.Enter):
				return ChangeState(taskListState, WithModelOption(
					newTaskListModel(mod),
				))
			case key.Matches(msg, Keys.Init):
				return func() tea.Msg {
					_, t, err := svc.Init(mod.ID)
					if err != nil {
						return newErrorMsg(err, "initializing module", "module", mod)
					}
					return ChangeState(taskState, WithModelOption(
						newTaskModel(t, mod, 0, 0),
					))()
				}
			case key.Matches(msg, Keys.Plan):
				return func() tea.Msg {
					_, t, err := runs.Create(currentWorkspace.ID, run.CreateOptions{
						PlanOnly: true,
					})
					if err != nil {
						return newErrorMsg(err, "creating run", "module", mod)
					}
					return ChangeState(taskState, WithModelOption(
						newTaskModel(t, mod, 0, 0),
					))()
				}
			case key.Matches(msg, Keys.ShowState):
				return func() tea.Msg {
					t, err := mod.ShowState(svc)
					if err != nil {
						return newErrorMsg(err, "creating show state task", "module", mod)
					}
					return ChangeState(taskState, WithModelOption(
						newTaskModel(t, mod, 0, 0),
					))()
				}
			}
		}

		return nil
	}

	return d
}
