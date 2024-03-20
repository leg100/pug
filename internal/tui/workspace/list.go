package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
	tasktui "github.com/leg100/pug/internal/tui/task"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

type ListMaker struct {
	WorkspaceService *workspace.Service
	ModuleService    *module.Service
	RunService       *run.Service
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tui.Model, error) {
	columns := tui.ParentColumns(tui.WorkspaceListKind, parent.Kind)
	columns = append(columns, table.IDColumn)

	cellsFunc := func(ws *workspace.Workspace) []table.Cell {
		cells := tui.ParentCells(tui.WorkspaceListKind, parent.Kind, ws.Resource)
		return append(cells, table.Cell{Str: string(ws.ID().String())})
	}
	table := table.New[*workspace.Workspace](columns, cellsFunc, width, height).
		WithSortFunc(workspace.Sort).
		WithParent(parent)

	return list{
		table:   table,
		svc:     m.WorkspaceService,
		modules: m.ModuleService,
		runs:    m.RunService,
		parent:  parent,
	}, nil
}

type list struct {
	table   table.Model[*workspace.Workspace]
	svc     *workspace.Service
	modules *module.Service
	runs    *run.Service
	parent  resource.Resource
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		var opts workspace.ListOptions
		if m.parent != resource.GlobalResource {
			opts.ModuleID = m.parent.ID()
		}
		return table.BulkInsertMsg[*workspace.Workspace](m.svc.List(opts))
	}
}

func (m list) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Enter):
			if ws, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.RunListKind, &ws.Resource)
			}
		case key.Matches(msg, tui.Keys.Init):
			return m, tasktui.TaskCmd(m.modules.Init, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, tui.Keys.Format):
			return m, tasktui.TaskCmd(m.modules.Format, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, tui.Keys.Validate):
			return m, tasktui.TaskCmd(m.modules.Validate, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, tui.Keys.Plan):
			return m, m.createRun(run.CreateOptions{})
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.Breadcrumbs("Workspaces", m.parent)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) Pagination() string {
	return ""
}

func (m list) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, tui.Keys.CloseHelp)
	return
}

func (m list) highlightedOrSelectedModuleIDs() []resource.ID {
	selected := maps.Values(m.table.HighlightedOrSelected())
	moduleIDs := make([]resource.ID, len(selected))
	for i, s := range selected {
		moduleIDs[i] = s.Module().ID()
	}
	return moduleIDs
}

func (m list) createRun(opts run.CreateOptions) tea.Cmd {
	// Handle the case where a user has pressed a key on an empty table with
	// zero rows
	if len(m.table.Items()) == 0 {
		return nil
	}

	// If items have been selected then clear the selection
	var deselectCmd tea.Cmd
	if len(m.table.Selected) > 0 {
		deselectCmd = tui.CmdHandler(table.DeselectMsg{})
	}

	cmd := func() tea.Msg {
		workspaces := m.table.HighlightedOrSelected()
		for workspaceID := range workspaces {
			_, err := m.runs.Create(workspaceID, opts)
			if err != nil {
				return tui.NewErrorMsg(err, "creating run")
			}
		}
		return tui.NavigationMsg(
			tui.Page{Kind: tui.RunListKind, Parent: m.parent},
		)
	}
	return tea.Batch(cmd, deselectCmd)
}
