package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

type workspaceListModelMaker struct {
	svc     *workspace.Service
	modules *module.Service
	runs    *run.Service
}

func (m *workspaceListModelMaker) makeModel(parent resource.Resource) (Model, error) {
	columns := parentColumns(WorkspaceListKind, parent.Kind)
	columns = append(columns, table.IDColumn)

	cellsFunc := func(ws *workspace.Workspace) []table.Cell {
		cells := parentCells(WorkspaceListKind, parent.Kind, ws.Resource)
		return append(cells, table.Cell{Str: string(ws.ID().String())})
	}
	table := table.New[*workspace.Workspace](columns).
		WithCellsFunc(cellsFunc).
		WithSortFunc(workspace.Sort).
		WithParent(parent)

	return workspaceListModel{
		table:   table,
		svc:     m.svc,
		modules: m.modules,
		runs:    m.runs,
		parent:  parent,
	}, nil
}

type workspaceListModel struct {
	table   table.Model[*workspace.Workspace]
	svc     *workspace.Service
	modules *module.Service
	runs    *run.Service
	parent  resource.Resource
}

func (m workspaceListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts workspace.ListOptions
		if m.parent != resource.GlobalResource {
			opts.ModuleID = m.parent.ID()
		}
		return table.BulkInsertMsg[*workspace.Workspace](m.svc.List(opts))
	}
}

func (m workspaceListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Enter):
			if ws, ok := m.table.Highlighted(); ok {
				return m, navigate(page{kind: RunListKind, resource: ws.Resource})
			}
		case key.Matches(msg, Keys.Init):
			return m, taskCmd(m.modules.Init, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, Keys.Format):
			return m, taskCmd(m.modules.Format, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, Keys.Validate):
			return m, taskCmd(m.modules.Validate, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, Keys.Plan):
			return m, m.createRun(run.CreateOptions{})
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m workspaceListModel) Title() string {
	return breadcrumbs("Workspaces", m.parent)
}

func (m workspaceListModel) View() string {
	return m.table.View()
}

func (m workspaceListModel) Pagination() string {
	return ""
}

func (m workspaceListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}

func (m workspaceListModel) highlightedOrSelectedModuleIDs() []resource.ID {
	selected := maps.Values(m.table.HighlightedOrSelected())
	moduleIDs := make([]resource.ID, len(selected))
	for i, s := range selected {
		moduleIDs[i] = s.Module().ID()
	}
	return moduleIDs
}

func (m workspaceListModel) createRun(opts run.CreateOptions) tea.Cmd {
	// Handle the case where a user has pressed a key on an empty table with
	// zero rows
	if len(m.table.Items()) == 0 {
		return nil
	}

	// If items have been selected then clear the selection
	var deselectCmd tea.Cmd
	if len(m.table.Selected) > 0 {
		deselectCmd = cmdHandler(table.DeselectMsg{})
	}

	cmd := func() tea.Msg {
		workspaces := m.table.HighlightedOrSelected()
		for workspaceID := range workspaces {
			_, err := m.runs.Create(workspaceID, opts)
			if err != nil {
				return newErrorMsg(err, "creating run")
			}
		}
		return navigationMsg{
			target: page{kind: RunListKind, resource: m.parent},
		}
	}
	return tea.Batch(cmd, deselectCmd)
}
