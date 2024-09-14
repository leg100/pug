package workspace

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
)

var currentColumn = table.Column{
	Key:   "current",
	Title: "CURRENT",
	Width: len("CURRENT"),
}

type ListMaker struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service
	Helpers    *tui.Helpers
}

func (m *ListMaker) Make(_ resource.ID, width, height int) (tea.Model, error) {
	columns := []table.Column{
		table.ModuleColumn,
		table.WorkspaceColumn,
		currentColumn,
		table.CostColumn,
		table.ResourceCountColumn,
	}

	renderer := func(ws *workspace.Workspace) table.RenderedRow {
		return table.RenderedRow{
			table.ModuleColumn.Key:        ws.ModulePath,
			table.WorkspaceColumn.Key:     ws.Name,
			table.ResourceCountColumn.Key: m.Helpers.WorkspaceResourceCount(ws),
			table.CostColumn.Key:          m.Helpers.WorkspaceCost(ws),
			currentColumn.Key:             m.Helpers.WorkspaceCurrentCheckmark(ws),
		}
	}

	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(workspace.Sort(m.Modules)),
	)

	return list{
		Workspaces: m.Workspaces,
		Modules:    m.Modules,
		Plans:      m.Plans,
		table:      table,
		Helpers:    m.Helpers,
	}, nil
}

type list struct {
	*tui.Helpers

	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service

	table table.Model[*workspace.Workspace]
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
		createRunOptions plan.CreateOptions
		applyPrompt      = "Auto-apply %d workspaces?"
		upgrade          bool
	)

	switch msg := msg.(type) {
	case resource.Event[*module.Module]:
		// Re-render workspaces belonging to updated module (the module's
		// current workspace may have changed, which changes the value of the
		// workspace's CURRENT column).
		workspaces := m.Workspaces.List(workspace.ListOptions{ModuleID: &msg.Payload.ID})
		for _, ws := range workspaces {
			m.table.AddItems(ws)
		}
	case resource.Event[*task.Task]:
		// Re-render workspace whenever a task event is received belonging to the
		// workspace.
		if workspaceID := msg.Payload.WorkspaceID; workspaceID != nil {
			ws, err := m.Workspaces.Get(*workspaceID)
			if err != nil {
				m.Logger.Error("re-rendering workspace upon receiving task event", "error", err, "task", msg.Payload)
				return m, nil
			}
			m.table.AddItems(ws)
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Delete):
			workspaceIDs := m.table.SelectedOrCurrentIDs()
			if len(workspaceIDs) == 0 {
				return m, nil
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Delete %d workspace(s)?", len(workspaceIDs)),
				m.CreateTasks(m.Workspaces.Delete, workspaceIDs...),
			)
		case key.Matches(msg, keys.Common.InitUpgrade):
			upgrade = true
			fallthrough
		case key.Matches(msg, keys.Common.Init):
			fn := func(moduleID resource.ID) (task.Spec, error) {
				return m.Modules.Init(moduleID, upgrade)
			}
			cmd := m.CreateTasks(fn, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := m.CreateTasks(m.Modules.Format, m.selectedOrCurrentModuleIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := m.CreateTasks(m.Modules.Validate, m.selectedOrCurrentModuleIDs()...)
			return m, cmd
		case key.Matches(msg, localKeys.SetCurrent):
			if row, ok := m.table.CurrentRow(); ok {
				return m, func() tea.Msg {
					if err := m.Workspaces.SelectWorkspace(row.Value.ModuleID, row.Value.ID); err != nil {
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
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Plan(workspaceID, createRunOptions)
			}
			return m, m.CreateTasks(fn, workspaceIDs...)
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			applyPrompt = "Destroy resources of %d workspaces?"
			fallthrough
		case key.Matches(msg, keys.Common.Apply):
			workspaceIDs := m.table.SelectedOrCurrentIDs()
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Apply(workspaceID, createRunOptions)
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(workspaceIDs)),
				m.CreateTasks(fn, workspaceIDs...),
			)
		case key.Matches(msg, keys.Common.State, localKeys.Enter):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(row.ID))
			}
		case key.Matches(msg, keys.Common.Cost):
			workspaceIDs := m.table.SelectedOrCurrentIDs()
			spec, err := m.Workspaces.Cost(workspaceIDs...)
			if err != nil {
				return m, tui.ReportError(fmt.Errorf("creating task: %w", err))
			}
			return m, m.CreateTasksWithSpecs(spec)
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return m.Breadcrumbs("Workspaces", nil)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Init,
		keys.Common.InitUpgrade,
		keys.Common.Format,
		keys.Common.Validate,
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.Apply,
		keys.Common.Destroy,
		keys.Common.Delete,
		keys.Common.Cost,
		localKeys.SetCurrent,
		keys.Common.State,
	}
}

// selectedOrCurrentModuleIDs returns the IDs of the modules of the
// current or selected workspaces.
func (m list) selectedOrCurrentModuleIDs() []resource.ID {
	selected := m.table.SelectedOrCurrent()
	moduleIDs := make([]resource.ID, len(selected))
	for i, row := range selected {
		moduleIDs[i] = row.Value.ModuleID
	}
	return moduleIDs
}
