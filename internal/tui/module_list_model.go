package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/common"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

//	func init() {
//		registerHelpBindings(func(short bool, current Page) []key.Binding {
//			if current != moduleListState {
//				return []key.Binding{Keys.Modules}
//			}
//			return []key.Binding{
//				Keys.Init,
//				Keys.Plan,
//				Keys.Apply,
//				Keys.ShowState,
//				Keys.Tasks,
//			}
//		})
//	}

type moduleListModelMaker struct {
	svc        *module.Service
	workspaces *workspace.Service
	workdir    string
}

func (m *moduleListModelMaker) makeModel(_ resource.Resource) (common.Model, error) {
	columns := []table.Column{
		table.NewFlexColumn(common.ColKeyPath, "PATH", 1).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(common.ColKeyID, "ID", resource.IDEncodedMaxLen).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}
	return moduleListModel{
		table:      table.New(columns).Focused(true).WithFooterVisibility(false),
		svc:        m.svc,
		modules:    make(map[resource.ID]*module.Module),
		workspaces: m.workspaces,
		workdir:    m.workdir,
	}, nil
}

type moduleListModel struct {
	table      table.Model
	svc        *module.Service
	workspaces *workspace.Service
	modules    map[resource.ID]*module.Module

	workdir string
}

func (mlm moduleListModel) Init() tea.Cmd {
	//return mlm.reload()
	return func() tea.Msg {
		return common.BulkInsertMsg[*module.Module](mlm.svc.List())
	}
}

func (m moduleListModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Enter):
			row := m.table.HighlightedRow()
			mod := row.Data[common.ColKeyData].(*module.Module)
			return m, navigate(page{kind: WorkspaceListKind, resource: mod.Resource})
		case key.Matches(msg, common.Keys.Init):
			row := m.table.HighlightedRow()
			mod := row.Data[common.ColKeyData].(*module.Module)
			return m, initCmd(m.svc, mod.ID)
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table = m.table.WithTargetWidth(msg.Width - 2)
		m.table = m.table.WithMinimumHeight(msg.Height)
	case common.BulkInsertMsg[*module.Module]:
		m.modules = make(map[resource.ID]*module.Module, len(msg))
		for _, mod := range msg {
			m.modules[mod.ID] = mod
		}
		m.table = m.table.WithRows(m.toRows())
	case resource.Event[*module.Module]:
		switch msg.Type {
		case resource.CreatedEvent:
			m.modules[msg.Payload.ID] = msg.Payload
		case resource.UpdatedEvent:
			m.modules[msg.Payload.ID] = msg.Payload
		case resource.DeletedEvent:
			delete(m.modules, msg.Payload.ID)
		}
		m.table = m.table.WithRows(m.toRows())
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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

func (m moduleListModel) toRows() []table.Row {
	rows := make([]table.Row, len(m.modules))
	for i, mod := range maps.Values(m.modules) {
		rows[i] = table.NewRow(table.RowData{
			common.ColKeyID:   mod.ID.String(),
			common.ColKeyPath: mod.Module().String(),
			common.ColKeyData: mod,
		})
	}
	return rows
}

func (mlm moduleListModel) reload() tea.Cmd {
	return func() tea.Msg {
		if err := mlm.svc.Reload(); err != nil {
			return common.NewErrorMsg(err, "reloading modules")
		}
		return nil
	}
}

func (mlm moduleListModel) reloadWorkspaces(module resource.Resource) tea.Cmd {
	return func() tea.Msg {
		_, err := mlm.workspaces.Reload(module)
		if err != nil {
			return common.NewErrorMsg(err, "reloading workspaces")
		}
		return nil
	}
}

func initCmd(modules *module.Service, moduleID resource.ID) tea.Cmd {
	return func() tea.Msg {
		_, task, err := modules.Init(moduleID)
		if err != nil {
			return common.NewErrorCmd(err, "creating init task")
		}
		return navigationMsg{
			target: page{kind: TaskKind, resource: task.Resource},
		}
	}
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
