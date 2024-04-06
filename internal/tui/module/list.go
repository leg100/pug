package module

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
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
	initColumn = table.Column{
		Key:   "init",
		Title: "INIT",
		Width: 4,
	}
	formatColumn = table.Column{
		Key:   "format",
		Title: "FORMAT",
		Width: 6,
	}
	validColumn = table.Column{
		Key:   "valid",
		Title: "VALID",
		Width: 5,
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
		table.RunStatusColumn,
		table.RunChangesColumn,
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
			table.ModuleColumn.Key:     mod.Path,
			initColumn.Key:             boolToUnicode(mod.InitInProgress, mod.Initialized),
			formatColumn.Key:           boolToUnicode(mod.FormatInProgress, mod.Formatted),
			validColumn.Key:            boolToUnicode(mod.ValidationInProgress, mod.Valid),
			currentWorkspace.Key:       m.Helpers.CurrentWorkspaceName(mod.CurrentWorkspaceID),
			table.RunStatusColumn.Key:  m.Helpers.ModuleCurrentRunStatus(mod),
			table.RunChangesColumn.Key: m.Helpers.ModuleCurrentRunChanges(mod),
		}
	}
	table := table.NewResource(table.ResourceOptions[*module.Module]{
		Columns:  columns,
		Renderer: renderer,
		Height:   height,
		Width:    width,
		SortFunc: module.ByPath,
	})

	return list{
		table:            table,
		spinner:          m.Spinner,
		ModuleService:    m.ModuleService,
		WorkspaceService: m.WorkspaceService,
		RunService:       m.RunService,
		workdir:          m.Workdir,
	}, nil
}

type list struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService

	table   table.Resource[resource.ID, *module.Module]
	spinner *spinner.Model
	workdir string
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[*module.Module](m.ModuleService.List())
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Handle keys that don't rely on any modules being present
		switch {
		case key.Matches(msg, localKeys.Reload):
			return m, tui.ReloadModules(m.WorkspaceService)
		}

		// Only handle following keys if there are modules present
		//
		// TODO: don't do this.
		if len(m.table.Items()) == 0 {
			break
		}

		switch {
		case key.Matches(msg, keys.Global.Enter):
			row, _ := m.table.Highlighted()
			return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(row.Value.Resource))
		case key.Matches(msg, localKeys.Reload):
			return m, tui.ReloadModules(m.WorkspaceService)
		case key.Matches(msg, localKeys.Edit):
			row, _ := m.table.Highlighted()
			return m, tui.OpenVim(row.Value.Path)
		case key.Matches(msg, localKeys.Init):
			cmd := tui.CreateTasks("init", m.ModuleService.Init, m.table.HighlightedOrSelectedKeys()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Validate):
			cmd := tui.CreateTasks("validate", m.ModuleService.Validate, m.table.HighlightedOrSelectedKeys()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Format):
			cmd := tui.CreateTasks("format", m.ModuleService.Format, m.table.HighlightedOrSelectedKeys()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Plan):
			workspaceIDs := m.HighlightedOrSelectedCurrentWorkspaceIDs()
			m.table.DeselectAll()
			return m, tui.CreateRuns(m.RunService, workspaceIDs...)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.GlobalBreadcrumb("Modules")
}

func (m list) View() string {
	return m.table.View()
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		localKeys.Init,
		localKeys.Validate,
		localKeys.Format,
		localKeys.Reload,
		localKeys.Plan,
		localKeys.Edit,
	}
}

// HighlightedOrSelectedCurrentWorkspaceIDs returns the IDs of the current
// workspaces of any modules that are currently selected or highlighted.
func (m list) HighlightedOrSelectedCurrentWorkspaceIDs() (workspaceIDs []resource.ID) {
	for _, row := range m.table.HighlightedOrSelected() {
		if current := row.Value.CurrentWorkspaceID; current != nil {
			workspaceIDs = append(workspaceIDs, *current)
		}
	}
	return
}
