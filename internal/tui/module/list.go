package module

import (
	"errors"
	"fmt"

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
	initColumn = table.Column{
		Key:   "init",
		Title: "INIT",
		Width: len("INIT"),
	}
	formatColumn = table.Column{
		Key:   "format",
		Title: "FORMAT",
		Width: len("FORMAT"),
	}
	validColumn = table.Column{
		Key:   "valid",
		Title: "VALID",
		Width: len("VALID"),
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
}

func (m *ListMaker) Make(_ resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{
		table.ModuleColumn,
		currentWorkspace,
		table.ResourceCountColumn,
		initColumn,
		formatColumn,
		validColumn,
	}
	boolToUnicode := func(inprog bool, b *bool) string {
		if inprog {
			return m.Spinner.View()
		}
		if b == nil {
			return "-"
		}
		if *b {
			return "✓"
		} else {
			return "✗"
		}
	}

	renderer := func(mod *module.Module) table.RenderedRow {
		return table.RenderedRow{
			table.ModuleColumn.Key:        mod.Path,
			initColumn.Key:                boolToUnicode(mod.InitInProgress, mod.Initialized),
			formatColumn.Key:              boolToUnicode(mod.FormatInProgress, mod.Formatted),
			validColumn.Key:               boolToUnicode(mod.ValidationInProgress, mod.Valid),
			currentWorkspace.Key:          m.Helpers.CurrentWorkspaceName(mod.CurrentWorkspaceID),
			table.ResourceCountColumn.Key: m.Helpers.ModuleCurrentResourceCount(mod),
		}
	}
	table := table.New(columns, renderer, width, height).WithSortFunc(module.ByPath)

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

	table   table.Model[resource.ID, *module.Module]
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
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.ReloadModules):
			return m, ReloadModules(false, m.ModuleService)
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(row.Value))
			}
		case key.Matches(msg, keys.Common.Edit):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.OpenVim(row.Value.Path)
			}
		case key.Matches(msg, keys.Common.Init):
			cmd := m.helpers.CreateTasks("init", m.ModuleService.Init, m.table.SelectedOrCurrentKeys()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := m.helpers.CreateTasks("validate", m.ModuleService.Validate, m.table.SelectedOrCurrentKeys()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := m.helpers.CreateTasks("format", m.ModuleService.Format, m.table.SelectedOrCurrentKeys()...)
			return m, cmd
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			cmd := m.helpers.CreateTasks("reload-workspace", m.WorkspaceService.Reload, m.table.SelectedOrCurrentKeys()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			workspaceIDs, err := m.pruneModulesWithoutCurrentWorkspace()
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.RunService.Plan(workspaceID, createRunOptions)
			}
			return m, m.helpers.CreateTasks("plan", fn, workspaceIDs...)
		case key.Matches(msg, keys.Common.Apply):
			workspaceIDs, err := m.pruneModulesWithoutCurrentWorkspace()
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.RunService.ApplyOnly(workspaceID, createRunOptions)
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Auto-apply %d modules?", len(workspaceIDs)),
				m.helpers.CreateTasks("apply", fn, workspaceIDs...),
			)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m list) pruneModulesWithoutCurrentWorkspace() ([]resource.ID, error) {
	workspaceIDs, err := m.table.Prune(func(mod *module.Module) (resource.ID, error) {
		if workspaceID := mod.CurrentWorkspaceID; workspaceID != nil {
			return *workspaceID, nil
		}
		return resource.ID{}, errors.New("module does not have a current workspace")
	})
	if err != nil {
		return nil, err
	}
	return workspaceIDs, nil
}

func (m list) Title() string {
	return tui.GlobalBreadcrumb("Modules", m.table.TotalString())
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
		keys.Common.Apply,
		keys.Common.Destroy,
		keys.Common.Edit,
		localKeys.ReloadModules,
		localKeys.ReloadWorkspaces,
	}
}
