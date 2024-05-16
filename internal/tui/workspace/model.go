package workspace

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	tuirun "github.com/leg100/pug/internal/tui/run"
	tuitask "github.com/leg100/pug/internal/tui/task"
	"github.com/leg100/pug/internal/workspace"
)

const (
	runsTabTitle      = "runs"
	tasksTabTitle     = "tasks"
	resourcesTabTitle = "resources"
)

// Maker makes workspace models.
type Maker struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	StateService     tui.StateService
	RunService       tui.RunService
	TaskService      tui.TaskService

	RunListMaker  *tuirun.ListMaker
	TaskListMaker *tuitask.ListMaker

	Spinner *spinner.Model
	Helpers *tui.Helpers
}

func (mm *Maker) Make(ws resource.Resource, width, height int) (tea.Model, error) {
	rlm := &resourceListMaker{
		StateService: mm.StateService,
		RunService:   mm.RunService,
		Spinner:      mm.Spinner,
	}

	tabs := tui.NewTabSet(width, height)
	_, err := tabs.AddTab(mm.RunListMaker, ws, runsTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding runs tab: %w", err)
	}
	_, err = tabs.AddTab(mm.TaskListMaker, ws, tasksTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding tasks tab: %w", err)
	}
	_, err = tabs.AddTab(rlm, ws, resourcesTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding resources tab: %w", err)
	}

	m := model{
		modules:   mm.ModuleService,
		runs:      mm.RunService,
		workspace: ws,
		tabs:      tabs,
		helpers:   mm.Helpers,
	}
	return m, nil
}

type model struct {
	modules   tui.ModuleService
	runs      tui.RunService
	workspace resource.Resource
	tabs      tui.TabSet
	helpers   *tui.Helpers
}

func (m model) Init() tea.Cmd {
	return m.tabs.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			// Create a plan. If the resources tab is selected, then ignore and
			// let the resources model handle creating a *targeted* plan.
			if m.tabs.ActiveTitle() == resourcesTabTitle {
				break
			}
			return m, tuirun.CreateRuns(m.runs, m.workspace, createRunOptions, m.workspace.GetID())
		case key.Matches(msg, keys.Common.Init):
			// create init task and switch user to its task page
			return m, func() tea.Msg {
				task, err := m.modules.Init(m.workspace.Module().GetID())
				if err != nil {
					return tui.NewErrorMsg(err, "creating init task")
				}
				return tui.NewNavigationMsg(tui.TaskKind, tui.WithParent(task))
			}
		case key.Matches(msg, keys.Common.Validate):
			return m, tuitask.CreateTasks("validate", m.workspace, m.modules.Validate, m.workspace.Module().GetID())
		case key.Matches(msg, keys.Common.Format):
			return m, tuitask.CreateTasks("format", m.workspace, m.modules.Format, m.workspace.Module().GetID())
		case key.Matches(msg, keys.Common.Module):
			return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(m.workspace.Module()))
		}
	case resource.Event[*workspace.Workspace]:
		if msg.Payload.ID == m.workspace.GetID() {
			m.workspace = msg.Payload
		}
	}
	// Navigate tabs
	updated, cmd := m.tabs.Update(msg)
	cmds = append(cmds, cmd)
	m.tabs = updated

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return m.helpers.Breadcrumbs("Workspace", m.workspace)
}

func (m model) View() string {
	return m.tabs.View()
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return append(
		m.tabs.HelpBindings(),
		keys.Common.Plan,
		keys.Common.Destroy,
		keys.Common.Init,
		keys.Common.Format,
		keys.Common.Validate,
		keys.Common.Module,
	)
}
