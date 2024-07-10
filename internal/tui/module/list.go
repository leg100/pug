package module

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
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
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService
	Spinner          *spinner.Model
	Workdir          string
	Helpers          *tui.Helpers
	Terragrunt       bool
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
		dependencyNames := make([]string, len(mod.Dependencies()))
		for i, id := range mod.Dependencies() {
			mod, _ := m.ModuleService.Get(id)
			dependencyNames[i] = mod.Path
		}
		row[dependencies.Key] = strings.Join(dependencyNames, ",")
		return row
	}
	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(module.ByPath),
	)

	return list{
		table:            table,
		spinner:          m.Spinner,
		ModuleService:    m.ModuleService,
		WorkspaceService: m.WorkspaceService,
		RunService:       m.RunService,
		workdir:          m.Workdir,
		helpers:          m.Helpers,
	}, nil
}

type list struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService

	table   table.Model[*module.Module]
	spinner *spinner.Model
	workdir string
	helpers *tui.Helpers
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[*module.Module](m.ModuleService.List())
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
		applyPrompt      = "Auto-apply %d modules?"
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.ReloadModules):
			return m, ReloadModules(false, m.ModuleService)
		case key.Matches(msg, keys.Common.Edit):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.OpenEditor(row.Value.FullPath())
			}
		case key.Matches(msg, keys.Common.Init):
			cmd := m.helpers.CreateTasks("init", m.ModuleService.Init, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := m.helpers.CreateTasks("validate", m.ModuleService.Validate, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := m.helpers.CreateTasks("format", m.ModuleService.Format, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			cmd := m.helpers.CreateTasks("reload-workspace", m.WorkspaceService.Reload, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.State):
			if row, ok := m.table.CurrentRow(); ok {
				if ws := m.helpers.ModuleCurrentWorkspace(row.Value); ws != nil {
					return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws))
				}
			}
		case key.Matches(msg, keys.Common.PlanDestroy):
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			workspaceIDs, err := m.pruneModulesWithoutCurrentWorkspace()
			if err != nil {
				return m, tui.ReportError(fmt.Errorf("deselected items: %w", err))
			}
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.RunService.Plan(workspaceID, createRunOptions)
			}
			desc := run.PlanTaskDescription(createRunOptions.Destroy)
			return m, m.helpers.CreateTasks(desc, fn, workspaceIDs...)
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			applyPrompt = "Destroy resources of %d modules?"
			fallthrough
		case key.Matches(msg, keys.Common.Apply):
			workspaceIDs, err := m.pruneModulesWithoutCurrentWorkspace()
			if err != nil {
				return m, tui.ReportError(fmt.Errorf("deselected items: %w", err))
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(workspaceIDs)),
				m.helpers.CreateApplyTasks(&createRunOptions, workspaceIDs...),
			)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *list) pruneModulesWithoutCurrentWorkspace() ([]resource.ID, error) {
	workspaceIDs, err := m.table.Prune(func(mod *module.Module) (resource.ID, bool) {
		if workspaceID := mod.CurrentWorkspaceID; workspaceID != nil {
			return *workspaceID, false
		}
		return resource.ID{}, true
	})
	if err != nil {
		return nil, errors.New("module(s) do not have a current workspace")
	}
	return workspaceIDs, nil
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
