package workspace

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	tuirun "github.com/leg100/pug/internal/tui/run"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
)

var currentColumn = table.Column{
	Key:        "current",
	Title:      "CURRENT",
	Width:      len("CURRENT"),
	FlexFactor: 1,
}

var resourceCountColumn = table.Column{
	Key:        "resource_count",
	Title:      "RESOURCES",
	Width:      len("RESOURCES"),
	FlexFactor: 1,
}

type ListMaker struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService
	Helpers          *tui.Helpers
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	var columns []table.Column
	if parent.Kind == resource.Global {
		// Show module column in global workspaces table
		columns = append(columns, table.ModuleColumn)
	}
	columns = append(columns,
		table.WorkspaceColumn,
		currentColumn,
		resourceCountColumn,
		table.RunStatusColumn,
		table.RunChangesColumn,
	)

	rowRenderer := func(ws *workspace.Workspace) table.RenderedRow {
		return table.RenderedRow{
			table.ModuleColumn.Key:     m.Helpers.ModulePath(ws.Resource),
			table.WorkspaceColumn.Key:  ws.Name,
			table.RunStatusColumn.Key:  m.Helpers.WorkspaceCurrentRunStatus(ws),
			table.RunChangesColumn.Key: m.Helpers.WorkspaceCurrentRunChanges(ws),
			resourceCountColumn.Key:    m.Helpers.WorkspaceResourceCount(ws),
			currentColumn.Key:          m.Helpers.WorkspaceCurrentCheckmark(ws),
		}
	}

	table := table.NewResource(table.ResourceOptions[*workspace.Workspace]{
		Columns:  columns,
		Renderer: rowRenderer,
		Width:    width,
		Height:   height,
		Parent:   parent,
		SortFunc: workspace.Sort(m.ModuleService),
	})

	return list{
		table:   table,
		svc:     m.WorkspaceService,
		modules: m.ModuleService,
		runs:    m.RunService,
		parent:  parent.ID,
	}, nil
}

type list struct {
	table   table.Resource[resource.ID, *workspace.Workspace]
	svc     tui.WorkspaceService
	modules tui.ModuleService
	runs    tui.RunService
	parent  resource.ID
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		workspaces := m.svc.List(workspace.ListOptions{
			ModuleID: m.parent,
		})
		return table.BulkInsertMsg[*workspace.Workspace](workspaces)
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
	)

	switch msg := msg.(type) {
	case resource.Event[*module.Module]:
		// Update changes to current workspace for a module
		m.table.UpdateViewport()
	case resource.Event[*run.Run]:
		// Update current run status and changes
		m.table.UpdateViewport()
	case resource.Event[*state.State]:
		// Update resource counts
		m.table.UpdateViewport()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Enter):
			if row, highlighted := m.table.Highlighted(); highlighted {
				return m, tui.NavigateTo(tui.WorkspaceKind, tui.WithParent(row.Value.Resource))
			}
		case key.Matches(msg, keys.Common.Delete):
			workspaceIDs := m.table.HighlightedOrSelectedKeys()
			if len(workspaceIDs) == 0 {
				return m, nil
			}
			return m, tui.RequestConfirmation(
				fmt.Sprintf("Delete %d workspace(s)", len(workspaceIDs)),
				tui.CreateTasks("delete-workspace", m.svc.Delete, workspaceIDs...),
			)
		case key.Matches(msg, keys.Common.Init):
			cmd := tui.CreateTasks("init", m.modules.Init, m.highlightedOrSelectedModuleIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := tui.CreateTasks("format", m.modules.Format, m.highlightedOrSelectedModuleIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := tui.CreateTasks("validate", m.modules.Validate, m.highlightedOrSelectedModuleIDs()...)
			return m, cmd
		case key.Matches(msg, localKeys.SetCurrent):
			if row, highlighted := m.table.Highlighted(); highlighted {
				return m, func() tea.Msg {
					if err := m.svc.SelectWorkspace(row.Value.ModuleID(), row.Value.ID); err != nil {
						return tui.NewErrorMsg(err, "setting current workspace")
					}
					return nil
				}
			}
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			workspaceIDs := m.table.HighlightedOrSelectedKeys()
			return m, tuirun.CreateRuns(m.runs, createRunOptions, workspaceIDs...)
		case key.Matches(msg, keys.Common.Apply):
			runIDs, err := m.table.Prune(func(ws *workspace.Workspace) (resource.ID, error) {
				if runID := ws.CurrentRunID; runID != nil {
					return *runID, nil
				}
				return resource.ID{}, errors.New("workspace does not have a current run")
			})
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			return m, tuirun.ApplyCommand(m.runs, runIDs...)
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.GlobalBreadcrumb("Workspaces")
}

func (m list) View() string {
	return m.table.View()
}

func (m list) TabStatus() string {
	return fmt.Sprintf("(%d)", len(m.table.Items()))
}

func (m list) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Init,
		keys.Common.Format,
		keys.Common.Validate,
		keys.Common.Plan,
		keys.Common.Apply,
		keys.Common.Destroy,
		localKeys.SetCurrent,
	}
}

// highlightedOrSelectedModuleIDs returns the IDs of the modules of the
// highlighted/selected workspaces.
func (m list) highlightedOrSelectedModuleIDs() []resource.ID {
	selected := m.table.HighlightedOrSelected()
	moduleIDs := make([]resource.ID, len(selected))
	for i, row := range selected {
		moduleIDs[i] = row.Value.ModuleID()
	}
	return moduleIDs
}
