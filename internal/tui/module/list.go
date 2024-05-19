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
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	tuirun "github.com/leg100/pug/internal/tui/run"
	"github.com/leg100/pug/internal/tui/table"
	tuitask "github.com/leg100/pug/internal/tui/task"
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
			table.ModuleColumn.Key:        mod.Path,
			initColumn.Key:                boolToUnicode(mod.InitInProgress, mod.Initialized),
			formatColumn.Key:              boolToUnicode(mod.FormatInProgress, mod.Formatted),
			validColumn.Key:               boolToUnicode(mod.ValidationInProgress, mod.Valid),
			currentWorkspace.Key:          m.Helpers.CurrentWorkspaceName(mod.CurrentWorkspaceID),
			table.RunStatusColumn.Key:     m.Helpers.ModuleCurrentRunStatus(mod),
			table.RunChangesColumn.Key:    m.Helpers.ModuleCurrentRunChanges(mod),
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
	}, nil
}

type list struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService

	table   table.Model[resource.ID, *module.Module]
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
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
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
		case key.Matches(msg, localKeys.ReloadModules):
			return m, ReloadModules(false, m.ModuleService)
		case key.Matches(msg, keys.Global.Enter):
			if row, highlighted := m.table.Highlighted(); highlighted {
				return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(row.Value))
			}
		case key.Matches(msg, keys.Common.Edit):
			if row, highlighted := m.table.Highlighted(); highlighted {
				return m, tui.OpenVim(row.Value.Path)
			}
		case key.Matches(msg, keys.Common.Init):
			rows := m.table.HighlightedOrSelected()
			switch len(rows) {
			case 0:
				// no rows, do nothing
			case 1:
				// create init task and switch user to its task page
				return m, func() tea.Msg {
					task, err := m.ModuleService.Init(rows[0].Key)
					if err != nil {
						return tui.NewErrorMsg(err, "creating init task")
					}
					return tui.NewNavigationMsg(tui.TaskKind, tui.WithParent(task))
				}
			default:
				// create init tasks, and keep user on current page.
				cmd := tuitask.CreateTasks("init", resource.GlobalResource, m.ModuleService.Init, m.table.HighlightedOrSelectedKeys()...)
				return m, cmd
			}
		case key.Matches(msg, keys.Common.Validate):
			cmd := tuitask.CreateTasks("validate", resource.GlobalResource, m.ModuleService.Validate, m.table.HighlightedOrSelectedKeys()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			cmd := tuitask.CreateTasks("format", resource.GlobalResource, m.ModuleService.Format, m.table.HighlightedOrSelectedKeys()...)
			return m, cmd
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			cmd := tuitask.CreateTasks("reload-workspace", resource.GlobalResource, m.WorkspaceService.Reload, m.table.HighlightedOrSelectedKeys()...)
			return m, cmd
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			workspaceIDs, err := m.table.Prune(func(mod *module.Module) (resource.ID, error) {
				if workspaceID := mod.CurrentWorkspaceID; workspaceID != nil {
					return *workspaceID, nil
				}
				return resource.ID{}, errors.New("module does not have a current workspace")
			})
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			return m, tuirun.CreateRuns(m.RunService, resource.GlobalResource, createRunOptions, workspaceIDs...)
		case key.Matches(msg, keys.Common.Apply):
			runIDs, err := m.table.Prune(func(mod *module.Module) (resource.ID, error) {
				curr, err := m.currentRun(mod)
				if err != nil {
					return resource.ID{}, err
				}
				if curr == nil {
					return resource.ID{}, errors.New("module does not have a current run")
				}
				if curr.Status != run.Planned {
					return resource.ID{}, fmt.Errorf("run not in unapplyable state: %s", string(curr.Status))
				}
				// Module has a current run and it is applyable, so do not
				// prune
				return curr.ID, nil
			})
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			return m, tuirun.ApplyCommand(m.RunService, resource.GlobalResource, runIDs...)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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

// currentRun retrieves current run for a module. If there is no current run for
// the module, then a nil run is returned.
func (m list) currentRun(mod *module.Module) (*run.Run, error) {
	currentWorkspaceID := mod.CurrentWorkspaceID
	if currentWorkspaceID == nil {
		return nil, nil
	}
	run, err := m.WorkspaceService.Get(*currentWorkspaceID)
	if err != nil {
		return nil, err
	}
	currentRunID := run.CurrentRunID
	if currentRunID == nil {
		return nil, nil
	}
	return m.RunService.Get(*currentRunID)
}
