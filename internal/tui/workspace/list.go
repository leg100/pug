package workspace

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

var currentColumn = table.Column{
	Key:   "current",
	Title: "CURRENT",
	// width of "CURRENT"; the actual content is a '✓' or nothing
	Width:      7,
	FlexFactor: 1,
}

type ListMaker struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tui.Model, error) {
	var columns []table.Column
	if parent.Kind() == resource.Global {
		columns = append(columns, table.ModuleColumn)
	}
	columns = append(columns,
		table.WorkspaceColumn,
		currentColumn,
		table.RunStatusColumn,
		table.RunChangesColumn,
	)

	table := table.New(columns, m.renderRow, width, height).
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

func (m *ListMaker) renderRow(ws *workspace.Workspace, inherit lipgloss.Style) table.RenderedRow {
	row := table.RenderedRow{
		table.ModuleColumn.Key:    ws.ModulePath(),
		table.WorkspaceColumn.Key: ws.Name(),
	}

	mod, _ := m.ModuleService.Get(ws.Module().ID())
	if mod.CurrentWorkspace != nil && *mod.CurrentWorkspace == ws.Resource {
		row[currentColumn.Key] = "✓"
	}

	if cr := ws.CurrentRun; cr != nil {
		run, _ := m.RunService.Get(cr.ID())
		row[table.RunStatusColumn.Key] = tui.RenderRunStatus(run.Status)
		row[table.RunChangesColumn.Key] = tui.RenderLatestRunReport(run, inherit)
	}
	return row
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
	case resource.Event[*workspace.Workspace]:
		// Update current workspace and current run
		m.table.UpdateViewport()
	case resource.Event[*run.Run]:
		// Update current run status and changes
		m.table.UpdateViewport()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Enter):
			if ws, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.WorkspaceKind, tui.WithParent(ws.Resource))
			}
		case key.Matches(msg, localKeys.Init):
			cmd := tui.CreateTasks("init", m.modules.Init, m.highlightedOrSelectedModuleIDs()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Format):
			cmd := tui.CreateTasks("format", m.modules.Format, m.highlightedOrSelectedModuleIDs()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Validate):
			cmd := tui.CreateTasks("validate", m.modules.Validate, m.highlightedOrSelectedModuleIDs()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Plan):
			workspaceIDs := m.table.HighlightedOrSelectedIDs()
			m.table.DeselectAll()
			return m, tui.CreateRuns(m.runs, workspaceIDs...)
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

func (m list) TabStatus() string {
	return fmt.Sprintf("(%d)", len(m.table.Items()))
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return keys.KeyMapToSlice(localKeys)
}

func (m list) highlightedOrSelectedModuleIDs() []resource.ID {
	selected := maps.Values(m.table.HighlightedOrSelected())
	moduleIDs := make([]resource.ID, len(selected))
	for i, s := range selected {
		moduleIDs[i] = s.Module().ID()
	}
	return moduleIDs
}
