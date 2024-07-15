package workspace

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
)

var currentColumn = table.Column{
	Key:        "current",
	Title:      "CURRENT",
	Width:      len("CURRENT"),
	FlexFactor: 1,
}

type ListMaker struct {
	Modules    tui.ModuleService
	Workspaces tui.WorkspaceService
	Runs       tui.RunService
	Helpers    *tui.Helpers
}

func (m *ListMaker) Make(_ resource.ID, width, height int) (tea.Model, error) {
	columns := []table.Column{
		table.ModuleColumn,
		table.WorkspaceColumn,
		currentColumn,
		table.ResourceCountColumn,
	}

	renderer := func(ws *workspace.Workspace) table.RenderedRow {
		return table.RenderedRow{
			table.ModuleColumn.Key:        ws.ModulePath(),
			table.WorkspaceColumn.Key:     ws.Name,
			table.ResourceCountColumn.Key: m.Helpers.WorkspaceResourceCount(ws),
			currentColumn.Key:             m.Helpers.WorkspaceCurrentCheckmark(ws),
		}
	}

	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(workspace.Sort(m.Modules)),
	)

	return list{
		Workspaces: m.Workspaces,
		Modules:    m.Modules,
		Runs:       m.Runs,
		table:      table,
		helpers:    m.Helpers,
	}, nil
}

type list struct {
	Modules    tui.ModuleService
	Workspaces tui.WorkspaceService
	Runs       tui.RunService

	table table.Model[*workspace.Workspace]

	helpers *tui.Helpers
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		workspaces := m.Workspaces.List(workspace.ListOptions{})
		return table.BulkInsertMsg[*workspace.Workspace](workspaces)
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
		applyPrompt      = "Auto-apply %d workspaces?"
	)

	switch msg := msg.(type) {
	case resource.Event[*module.Module]:
		// Update changes to current workspace for a module
		m.table.UpdateViewport()
	case resource.Event[*state.State]:
		// Update resource counts
		m.table.UpdateViewport()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Delete):
			workspaceIDs := m.table.SelectedOrCurrentIDs()
			if len(workspaceIDs) == 0 {
				return m, nil
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Delete %d workspace(s)?", len(workspaceIDs)),
				m.helpers.CreateTasks("delete-workspace", m.Workspaces.Delete, workspaceIDs...),
			)
		case key.Matches(msg, keys.Common.Init):
			cmd := m.helpers.CreateTasks("init", m.Modules.Init, m.selectedOrCurrentModuleIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := m.helpers.CreateTasks("format", m.Modules.Format, m.selectedOrCurrentModuleIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := m.helpers.CreateTasks("validate", m.Modules.Validate, m.selectedOrCurrentModuleIDs()...)
			return m, cmd
		case key.Matches(msg, localKeys.SetCurrent):
			if row, ok := m.table.CurrentRow(); ok {
				return m, func() tea.Msg {
					if err := m.Workspaces.SelectWorkspace(row.Value.ModuleID(), row.Value.ID); err != nil {
						return tui.ReportError(fmt.Errorf("setting current workspace: %w", err))()
					}
					return nil
				}
			}
		case key.Matches(msg, keys.Common.PlanDestroy):
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			workspaceIDs := m.table.SelectedOrCurrentIDs()
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.Runs.Plan(workspaceID, createRunOptions)
			}
			desc := run.PlanTaskDescription(createRunOptions.Destroy)
			return m, m.helpers.CreateTasks(desc, fn, workspaceIDs...)
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			applyPrompt = "Destroy resources of %d workspaces?"
			fallthrough
		case key.Matches(msg, keys.Common.Apply):
			workspaceIDs := m.table.SelectedOrCurrentIDs()
			return m, tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(workspaceIDs)),
				m.helpers.CreateApplyTasks(&createRunOptions, workspaceIDs...),
			)
		case key.Matches(msg, keys.Common.State):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(row.Value))
			}
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.Breadcrumbs("Workspaces", resource.GlobalResource)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Init,
		keys.Common.Format,
		keys.Common.Validate,
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.Apply,
		keys.Common.Destroy,
		keys.Common.Delete,
		localKeys.SetCurrent,
	}
}

// selectedOrCurrentModuleIDs returns the IDs of the modules of the
// current or selected workspaces.
func (m list) selectedOrCurrentModuleIDs() []resource.ID {
	selected := m.table.SelectedOrCurrent()
	moduleIDs := make([]resource.ID, len(selected))
	for i, row := range selected {
		moduleIDs[i] = row.Value.ModuleID()
	}
	return moduleIDs
}
