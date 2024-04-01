package workspace

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	runtui "github.com/leg100/pug/internal/tui/run"
	tasktui "github.com/leg100/pug/internal/tui/task"
	"github.com/leg100/pug/internal/workspace"
)

const (
	runsTabTitle      = "runs"
	tasksTabTitle     = "tasks"
	resourcesTabTitle = "resources"
)

// Maker makes workspace models.
type Maker struct {
	WorkspaceService *workspace.Service
	StateService     *state.Service
	RunService       *run.Service
	TaskService      *task.Service

	RunListMaker  *runtui.ListMaker
	TaskListMaker *tasktui.ListMaker

	Spinner *spinner.Model
	Helpers *tui.Helpers
}

func (mm *Maker) Make(workspace resource.Resource, width, height int) (tui.Model, error) {
	ws, err := mm.WorkspaceService.Get(workspace.ID)
	if err != nil {
		return model{}, err
	}
	rlm := &resourceListMaker{
		StateService: mm.StateService,
		Spinner:      mm.Spinner,
	}

	tabs := tui.NewTabSet(width, height)
	_, err = tabs.AddTab(mm.RunListMaker, workspace, runsTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding runs tab: %w", err)
	}
	_, err = tabs.AddTab(mm.TaskListMaker, workspace, tasksTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding tasks tab: %w", err)
	}
	_, err = tabs.AddTab(rlm, workspace, resourcesTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding resources tab: %w", err)
	}

	m := model{
		runs:      mm.RunService,
		workspace: ws,
		tabs:      tabs,
		helpers:   mm.Helpers,
	}
	return m, nil
}

type model struct {
	runs      *run.Service
	workspace *workspace.Workspace
	tabs      tui.TabSet
	helpers   *tui.Helpers
}

func (m model) Init() tea.Cmd {
	return m.tabs.Init()
}

func (m model) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Plan):
			return m, tui.CreateRuns(m.runs, m.workspace.ID)
		case key.Matches(msg, keys.Common.Module):
			return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(*m.workspace.Module()))
		}
	case resource.Event[*workspace.Workspace]:
		if msg.Payload.ID == m.workspace.ID {
			m.workspace = msg.Payload
		}
	}
	// Update tabs
	updated, cmd := m.tabs.Update(msg)
	cmds = append(cmds, cmd)
	m.tabs = updated

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return m.helpers.Breadcrumbs("Workspace", m.workspace.Resource)
}

func (m model) View() string {
	return m.tabs.View()
}

func (m model) Pagination() string {
	return ""
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Apply,
	}
}
