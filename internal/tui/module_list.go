package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
)

type moduleListModelMaker struct {
	svc        *module.Service
	workspaces *workspace.Service
	workdir    string
}

func (m *moduleListModelMaker) makeModel(_ resource.Resource) (Model, error) {
	columns := []table.Column{
		table.NewFlexColumn(ColKeyModule, "PATH", 1).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyWorkspace, "CURRENT WORKSPACE", 20).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyID, "ID", resource.IDEncodedMaxLen).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}
	rowMaker := func(mod *module.Module) table.RowData {
		return table.RowData{
			ColKeyID:        mod.ID().String(),
			ColKeyModule:    mod.Path(),
			ColKeyWorkspace: mod.Current,
		}
	}
	return moduleListModel{
		table: newTableModel(tableModelOptions[*module.Module]{
			rowMaker: rowMaker,
			columns:  columns,
		}),
		svc:        m.svc,
		workspaces: m.workspaces,
		workdir:    m.workdir,
	}, nil
}

type moduleListModel struct {
	table      tableModel[*module.Module]
	svc        *module.Service
	workspaces *workspace.Service

	workdir string
}

func (mlm moduleListModel) Init() tea.Cmd {
	return func() tea.Msg {
		return bulkInsertMsg[*module.Module](mlm.svc.List())
	}
}

func (m moduleListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Enter):
			if ok, mod := m.table.highlighted(); ok {
				return m, navigate(page{kind: WorkspaceListKind, resource: mod.Resource})
			}
		case key.Matches(msg, Keys.Init):
			// TODO: consider sending another msg to clear selections
			return m, taskCmd(m.svc.Init, m.table.highlightedOrSelectedIDs()...)
		case key.Matches(msg, Keys.Validate):
			return m, taskCmd(m.svc.Validate, m.table.highlightedOrSelectedIDs()...)
		case key.Matches(msg, Keys.Format):
			return m, taskCmd(m.svc.Format, m.table.highlightedOrSelectedIDs()...)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (mlm moduleListModel) Title() string {
	return lipgloss.NewStyle().
		Inherit(Breadcrumbs).
		Padding(0, 0, 0, 1).
		Render(
			fmt.Sprintf("modules (%s)", mlm.workdir),
		)
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

func (m moduleListModel) Pagination() string {
	return m.table.Pagination()
}

func (mlm moduleListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}

func (mlm moduleListModel) reload() tea.Cmd {
	return func() tea.Msg {
		if err := mlm.svc.Reload(); err != nil {
			return newErrorMsg(err, "reloading modules")
		}
		return nil
	}
}

func (mlm moduleListModel) reloadWorkspaces(module resource.Resource) tea.Cmd {
	return func() tea.Msg {
		_, err := mlm.workspaces.Reload(module)
		if err != nil {
			return newErrorMsg(err, "reloading workspaces")
		}
		return nil
	}
}
