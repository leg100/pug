package module

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
)

// ListMaker makes module list models
type ListMaker struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
	Spinner          *spinner.Model
	Workdir          string
}

func (m *ListMaker) Make(_ resource.Resource, width, height int) (tui.Model, error) {
	currentWorkspace := table.Column{
		Key:        "currentWorkspace",
		Title:      "CURRENT WORKSPACE",
		FlexFactor: 2,
	}
	initColumn := table.Column{
		Key:   "init",
		Title: "INIT",
		Width: 4,
	}
	formatColumn := table.Column{
		Key:   "format",
		Title: "FORMAT",
		Width: 6,
	}
	validColumn := table.Column{
		Key:   "valid",
		Title: "VALID",
		Width: 5,
	}
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

	renderer := func(mod *module.Module, inherit lipgloss.Style) table.RenderedRow {
		row := table.RenderedRow{
			table.ModuleColumn.Key: mod.Path(),
			initColumn.Key:         boolToUnicode(mod.InitInProgress, mod.Initialized),
			formatColumn.Key:       boolToUnicode(mod.FormatInProgress, mod.Formatted),
			validColumn.Key:        boolToUnicode(mod.ValidationInProgress, mod.Valid),
		}
		// Retrieve name of current workspace if module has one
		if current := mod.CurrentWorkspace; current != nil {
			row[currentWorkspace.Key] = current.String()

			// Retrieve current run if current workspace has one
			ws, _ := m.WorkspaceService.Get(current.ID())
			if cr := ws.CurrentRun; cr != nil {
				run, _ := m.RunService.Get(cr.ID())
				row[table.RunStatusColumn.Key] = tui.RenderRunStatus(run.Status)
				row[table.RunChangesColumn.Key] = tui.RenderLatestRunReport(run, inherit)
			}
		}
		return row
	}
	table := table.New(columns, renderer, width, height)
	table = table.WithSortFunc(module.ByPath)

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
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service

	table   table.Model[*module.Module]
	spinner *spinner.Model
	workdir string
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[*module.Module](m.ModuleService.List())
	}
}

func (m list) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case resource.Event[*workspace.Workspace]:
		switch msg.Type {
		case resource.CreatedEvent:
			//cmds = append(cmds, m.createRun(run.CreateOptions{}))
			// navigate to the resources tab on the newly created workspace
			if msg.Payload.ModulePath() == "internal/testdata/configs/envs/prod" &&
				msg.Payload.Name() == "staging" {

				return m, tui.NavigateTo(tui.WorkspaceKind,
					tui.WithParent(msg.Payload.Resource),
				)
			}
		}
	case resource.Event[*run.Run]:
		switch msg.Type {
		case resource.UpdatedEvent:
			if msg.Payload.Status == run.Planned {
				//return m, tui.NavigateTo(tui.RunKind, &msg.Payload.Resource)
			}
		}
	}

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
		if len(m.table.Items()) == 0 {
			break
		}

		switch {
		case key.Matches(msg, keys.Global.Enter):
			mod, _ := m.table.Highlighted()
			return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(mod.Resource))
		case key.Matches(msg, localKeys.Reload):
			return m, tui.ReloadModules(m.WorkspaceService)
		case key.Matches(msg, localKeys.Edit):
			mod, _ := m.table.Highlighted()
			return m, tui.OpenVim(mod.Path())
		case key.Matches(msg, localKeys.Init):
			cmd := tui.CreateTasks("init", m.ModuleService.Init, m.table.HighlightedOrSelectedIDs()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Validate):
			cmd := tui.CreateTasks("validate", m.ModuleService.Validate, m.table.HighlightedOrSelectedIDs()...)
			m.table.DeselectAll()
			return m, cmd
		case key.Matches(msg, localKeys.Format):
			cmd := tui.CreateTasks("format", m.ModuleService.Format, m.table.HighlightedOrSelectedIDs()...)
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
	workdir := tui.Regular.Copy().Foreground(tui.Blue).Render("all")
	modules := tui.Bold.Copy().Render("Modules")
	return fmt.Sprintf("%s(%s)", modules, workdir)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) Pagination() string {
	return ""
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
	for _, mod := range m.table.HighlightedOrSelected() {
		if current := mod.CurrentWorkspace; current != nil {
			workspaceIDs = append(workspaceIDs, current.ID())
		}
	}
	return
}
