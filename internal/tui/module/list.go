package module

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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

var (
	currentWorkspace = table.Column{
		Key:        "currentWorkspace",
		Title:      "CURRENT WORKSPACE",
		FlexFactor: 2,
	}
	backendType = table.Column{
		Key:   "backendType",
		Title: "BACKEND",
		Width: len("BACKEND"),
	}
	dependencies = table.Column{
		Key:        "moduleDependencies",
		Title:      "DEPENDENCIES",
		FlexFactor: 2,
	}
)

// ListMaker makes module list models
type ListMaker struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service
	Spinner    *spinner.Model
	Workdir    string
	Helpers    *tui.Helpers
	Terragrunt bool
}

func (m *ListMaker) Make(_ resource.ID, width, height int) (tea.Model, error) {
	columns := []table.Column{
		table.ModuleColumn,
	}
	// Only include dependencies column if using terragrunt
	if m.Terragrunt {
		columns = append(columns, dependencies)
	}
	columns = append(columns,
		backendType,
		currentWorkspace,
		table.ResourceCountColumn,
	)

	renderer := func(mod *module.Module) table.RenderedRow {
		row := table.RenderedRow{
			table.ModuleColumn.Key:        mod.Path,
			backendType.Key:               mod.Backend,
			currentWorkspace.Key:          m.Helpers.CurrentWorkspaceName(mod.CurrentWorkspaceID),
			table.ResourceCountColumn.Key: m.Helpers.ModuleCurrentResourceCount(mod),
		}
		dependencyNames := make([]string, 0, len(mod.Dependencies()))
		for _, id := range mod.Dependencies() {
			mod, err := m.Modules.Get(id)
			if err != nil {
				// Should never happen
				dependencyNames = append(dependencyNames, fmt.Sprintf("error: %s", err.Error()))
				continue
			}
			dependencyNames = append(dependencyNames, mod.Path)
		}
		row[dependencies.Key] = strings.Join(dependencyNames, ",")
		return row
	}
	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(module.ByPath),
	)

	return list{
		table:      table,
		spinner:    m.Spinner,
		Modules:    m.Modules,
		Workspaces: m.Workspaces,
		Plans:      m.Plans,
		workdir:    m.Workdir,
		helpers:    m.Helpers,
	}, nil
}

type list struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service

	table   table.Model[*module.Module]
	spinner *spinner.Model
	workdir string
	helpers *tui.Helpers
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[*module.Module](m.Modules.List())
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd            tea.Cmd
		cmds           []tea.Cmd
		createPlanOpts plan.CreateOptions
		applyPrompt    = "Auto-apply %d modules?"
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.ReloadModules):
			return m, ReloadModules(false, m.Modules)
		case key.Matches(msg, keys.Common.Edit):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.OpenEditor(row.Value.FullPath())
			}
		case key.Matches(msg, keys.Common.Init):
			cmd := m.helpers.CreateTasks(m.Modules.Init, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := m.helpers.CreateTasks(m.Modules.Validate, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := m.helpers.CreateTasks(m.Modules.Format, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			cmd := m.helpers.CreateTasks(m.Workspaces.Reload, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.State):
			if row, ok := m.table.CurrentRow(); ok {
				if ws := m.helpers.ModuleCurrentWorkspace(row.Value); ws != nil {
					return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws))
				}
			}
		case key.Matches(msg, keys.Common.PlanDestroy):
			createPlanOpts.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			// Create specs here, de-selecting any modules where an error is
			// returned.
			specs, err := m.table.Prune(func(mod *module.Module) (task.Spec, error) {
				if workspaceID := mod.CurrentWorkspaceID; workspaceID == nil {
					return task.Spec{}, fmt.Errorf("module %s does not have a current workspace", mod)
				} else {
					return m.Plans.Plan(*workspaceID, createPlanOpts)
				}
			})
			if err != nil {
				// Modules were de-selected, so report error and give user
				// another opportunity to plan any remaining modules.
				return m, tui.ReportError(err)
			}
			return m, m.helpers.CreateTasksWithSpecs(specs...)
		case key.Matches(msg, keys.Common.Destroy):
			createPlanOpts.Destroy = true
			applyPrompt = "Destroy resources of %d modules?"
			fallthrough
		case key.Matches(msg, keys.Common.Apply):
			// Create specs here, de-selecting any modules where an error is
			// returned.
			specs, err := m.table.Prune(func(mod *module.Module) (task.Spec, error) {
				if workspaceID := mod.CurrentWorkspaceID; workspaceID == nil {
					return task.Spec{}, fmt.Errorf("module %s does not have a current workspace", mod)
				} else {
					return m.Plans.Apply(*workspaceID, createPlanOpts)
				}
			})
			if err != nil {
				// Modules were de-selected, so report error and give user
				// another opportunity to apply any remaining modules.
				return m, tui.ReportError(err)
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(specs)),
				m.helpers.CreateTasksWithSpecs(specs...),
			)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.Breadcrumbs("Modules", resource.GlobalResource)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Init,
		keys.Common.Format,
		keys.Common.Validate,
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.Apply,
		keys.Common.Destroy,
		keys.Common.Edit,
		localKeys.ReloadModules,
		localKeys.ReloadWorkspaces,
	}
}
