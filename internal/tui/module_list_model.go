package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
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
		table.NewColumn(common.ColKeyWorkspace, "CURRENT WORKSPACE", 20).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(common.ColKeyID, "ID", resource.IDEncodedMaxLen).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}
	return moduleListModel{
		table:      newDefaultTable(columns...),
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
			var moduleIDs []resource.ID
			if selected := m.table.SelectedRows(); len(selected) > 0 {
				for _, s := range selected {
					mod := s.Data[common.ColKeyData].(*module.Module)
					moduleIDs = append(moduleIDs, mod.ID)
				}
			} else {
				// TODO: if there are zero items then this'll blow up...
				row := m.table.HighlightedRow()
				mod := row.Data[common.ColKeyData].(*module.Module)
				moduleIDs = append(moduleIDs, mod.ID)
			}
			return m, moduleCmd(func(id resource.ID) (*task.Task, error) {
				_, task, err := m.svc.Init(id)
				return task, err
			}, moduleIDs...)
		case key.Matches(msg, common.Keys.SelectAll):
			selected := make([]table.Row, len(m.modules))
			for i, r := range m.toRows() {
				selected[i] = r.Selected(true)
			}
			m.table = m.table.WithRows(selected)
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table = m.table.WithTargetWidth(msg.Width - 2)
		m.table = m.table.WithMinimumHeight(msg.Height)
	case common.BulkInsertMsg[*module.Module]:
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

func (mlm moduleListModel) Footer(width int) string {
	return "logs"
}

func (mlm moduleListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

func (m moduleListModel) toRows() []table.Row {
	rows := make([]table.Row, len(m.modules))
	for i, mod := range maps.Values(m.modules) {
		data := table.RowData{
			common.ColKeyID:        mod.ID.String(),
			common.ColKeyPath:      mod.Path(),
			common.ColKeyWorkspace: mod.Current,
			common.ColKeyData:      mod,
		}
		rows[i] = table.NewRow(data)
	}
	return rows
}

func toRows(mod *module.Module) table.RowData {
	return table.RowData{
		common.ColKeyID:        mod.ID.String(),
		common.ColKeyPath:      mod.Path(),
		common.ColKeyWorkspace: mod.Current,
		// this would be handled by tableModel using generics
		common.ColKeyData: mod,
	}
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

func moduleCmd(fn func(resource.ID) (*task.Task, error), moduleIDs ...resource.ID) tea.Cmd {
	return func() tea.Msg {
		var task *task.Task
		for _, id := range moduleIDs {
			var err error
			task, err = fn(id)
			if err != nil {
				return common.NewErrorCmd(err, "creating task")
			}
		}
		if len(moduleIDs) > 1 {
			return navigationMsg{
				target: page{kind: TaskListKind},
			}
		}
		return navigationMsg{
			target: page{kind: TaskKind, resource: task.Resource},
		}
	}
}
