package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/tui/common"
)

//func init() {
//	registerHelpBindings(func(short bool, current Page) []key.Binding {
//		if current != moduleListState {
//			return []key.Binding{Keys.Modules}
//		}
//		return []key.Binding{
//			Keys.Init,
//			Keys.Plan,
//			Keys.Apply,
//			Keys.ShowState,
//			Keys.Tasks,
//		}
//	})
//}

type moduleListModel struct {
	table table.Model

	workdir string

	width  int
	height int
}

func NewModuleListModel(svc *module.Service, workdir string) (moduleListModel, error) {
	if err := svc.Reload(); err != nil {
		return moduleListModel{}, err
	}
	mods := svc.List()
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "PATH", Width: 10},
	}
	rows := make([]table.Row, len(mods))
	for i, t := range mods {
		rows[i] = newModuleRow(t)
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	return moduleListModel{table: tbl, workdir: workdir}, nil
}

func newModuleRow(m *module.Module) table.Row {
	return table.Row{
		m.ID.String(),
		m.Path,
	}
}

func (moduleListModel) Init() tea.Cmd {
	return nil
}

func (lm moduleListModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		lm.table.SetWidth(msg.Width - 2)
		lm.table.SetHeight(msg.Height)
		return lm, nil
	}

	var cmd tea.Cmd
	lm.table, cmd = lm.table.Update(msg)
	return lm, cmd
}

func (mlm moduleListModel) Title() string {
	return fmt.Sprintf("modules (%s)", mlm.workdir)
}

func (mlm moduleListModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		mlm.table.View(),
	)
}

func (mlm moduleListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

//func newModuleDelegate(svc *module.Service, runs *run.Service, workspaces *workspace.Service) list.DefaultDelegate {
//	d := list.NewDefaultDelegate()
//	d.SetSpacing(0)
//	d.ShowDescription = false
//
//	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
//		mod, ok := m.SelectedItem().(*module.Module)
//		if !ok {
//			return nil
//		}
//		currentWorkspace, err := workspaces.GetCurrent(mod.ID)
//		if err != nil {
//			return cmdHandler(newErrorMsg(err, "creating init task", "module", mod))
//		}
//
//		switch msg := msg.(type) {
//		case tea.KeyMsg:
//			switch {
//			case key.Matches(msg, Keys.Tasks, Keys.Enter):
//				return navigate(taskListState, WithModelOption(
//					NewTaskListModel(mod),
//				))
//			case key.Matches(msg, Keys.Init):
//				return func() tea.Msg {
//					_, t, err := svc.Init(mod.ID)
//					if err != nil {
//						return newErrorMsg(err, "initializing module", "module", mod)
//					}
//					return navigate(taskState, WithModelOption(
//						newTaskModel(t, mod, 0, 0),
//					))()
//				}
//			case key.Matches(msg, Keys.Plan):
//				return func() tea.Msg {
//					_, t, err := runs.Create(currentWorkspace.ID, run.CreateOptions{
//						PlanOnly: true,
//					})
//					if err != nil {
//						return newErrorMsg(err, "creating run", "module", mod)
//					}
//					return navigate(taskState, WithModelOption(
//						newTaskModel(t, mod, 0, 0),
//					))()
//				}
//			case key.Matches(msg, Keys.ShowState):
//				return func() tea.Msg {
//					t, err := mod.ShowState(svc)
//					if err != nil {
//						return newErrorMsg(err, "creating show state task", "module", mod)
//					}
//					return navigate(taskState, WithModelOption(
//						newTaskModel(t, mod, 0, 0),
//					))()
//				}
//			}
//		}
//
//		return nil
//	}
//
//	return d
//}
