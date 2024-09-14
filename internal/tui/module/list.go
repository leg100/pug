package module

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
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
	Workdir    internal.Workdir
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
		Helpers:    m.Helpers,
	}, nil
}

type list struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service

	table   table.Model[*module.Module]
	spinner *spinner.Model
	workdir internal.Workdir

	*tui.Helpers
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
		upgrade        bool
	)

	switch msg := msg.(type) {
	case resource.Event[*task.Task]:
		// Re-render module whenever a task event is received belonging to the
		// module.
		if moduleID := msg.Payload.ModuleID; moduleID != nil {
			mod, err := m.Modules.Get(*moduleID)
			if err != nil {
				m.Logger.Error("re-rendering module upon receiving task event", "error", err, "task", msg.Payload)
				return m, nil
			}
			m.table.AddItems(mod)
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.ReloadModules):
			return m, ReloadModules(false, m.Modules)
		case key.Matches(msg, keys.Common.Edit):
			if row, ok := m.table.CurrentRow(); ok {
				path := m.workdir.Join(row.Value.Path)
				return m, tui.OpenEditor(path)
			}
		case key.Matches(msg, keys.Common.InitUpgrade):
			upgrade = true
			fallthrough
		case key.Matches(msg, keys.Common.Init):
			fn := func(moduleID resource.ID) (task.Spec, error) {
				return m.Modules.Init(moduleID, upgrade)
			}
			cmd := m.CreateTasks(fn, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := m.CreateTasks(m.Modules.Validate, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := m.CreateTasks(m.Modules.Format, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			cmd := m.CreateTasks(m.Workspaces.Reload, m.table.SelectedOrCurrentIDs()...)
			return m, cmd
		case key.Matches(msg, keys.Common.State, localKeys.Enter):
			row, ok := m.table.CurrentRow()
			if !ok {
				return m, nil
			}
			ws := m.ModuleCurrentWorkspace(row.Value)
			if ws == nil {
				return m, tui.ReportError(errors.New("module does not have a current workspace"))
			}
			return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws.ID))
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
			return m, m.CreateTasksWithSpecs(specs...)
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
				m.CreateTasksWithSpecs(specs...),
			)
		case key.Matches(msg, localKeys.Execute):
			ids := m.table.SelectedOrCurrentIDs()

			return m, tui.CmdHandler(tui.PromptMsg{
				Prompt:      fmt.Sprintf("Execute program in %d module directories: ", len(ids)),
				Placeholder: "terraform version",
				Action: func(v string) tea.Cmd {
					if v == "" {
						return nil
					}
					// split value into program and any args
					parts := strings.Split(v, " ")
					prog := parts[0]
					args := parts[1:]
					fn := func(moduleID resource.ID) (task.Spec, error) {
						return m.Modules.Execute(moduleID, prog, args...)
					}
					return m.CreateTasks(fn, ids...)
				},
				Key:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
				Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
			})
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return m.Breadcrumbs("Modules", nil)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Init,
		keys.Common.InitUpgrade,
		keys.Common.Format,
		keys.Common.Validate,
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.Apply,
		keys.Common.Destroy,
		keys.Common.Edit,
		localKeys.Execute,
		localKeys.ReloadModules,
		localKeys.ReloadWorkspaces,
		keys.Common.State,
	}
}
