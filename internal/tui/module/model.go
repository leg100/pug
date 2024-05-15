package module

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	tuirun "github.com/leg100/pug/internal/tui/run"
	tuitask "github.com/leg100/pug/internal/tui/task"
	tuiworkspace "github.com/leg100/pug/internal/tui/workspace"
)

const (
	workspacesTabTitle = "workspaces"
	runsTabTitle       = "runs"
	tasksTabTitle      = "tasks"
)

// Maker makes module models.
type Maker struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService

	WorkspaceListMaker *tuiworkspace.ListMaker
	RunListMaker       *tuirun.ListMaker
	TaskListMaker      *tuitask.ListMaker

	Helpers *tui.Helpers
}

func (mm *Maker) Make(mr resource.Resource, width, height int) (tea.Model, error) {
	mod, err := mm.ModuleService.Get(mr.GetID())
	if err != nil {
		return model{}, err
	}

	tabs := tui.NewTabSet(width, height)
	_, err = tabs.AddTab(mm.WorkspaceListMaker, mr, workspacesTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding workspaces tab: %w", err)
	}
	_, err = tabs.AddTab(mm.RunListMaker, mr, runsTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding runs tab: %w", err)
	}
	_, err = tabs.AddTab(mm.TaskListMaker, mr, tasksTabTitle)
	if err != nil {
		return nil, fmt.Errorf("adding tasks tab: %w", err)
	}

	m := model{
		ModuleService:    mm.ModuleService,
		WorkspaceService: mm.WorkspaceService,
		RunService:       mm.RunService,
		module:           mod,
		tabs:             tabs,
		helpers:          mm.Helpers,
	}
	return m, nil
}

type model struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService

	module  *module.Module
	tabs    tui.TabSet
	helpers *tui.Helpers
}

func (m model) Init() tea.Cmd {
	return m.tabs.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// General actions
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Init):
			return m, tuitask.CreateTasks("init", resource.GlobalResource, m.ModuleService.Init, m.module.ID)
		case key.Matches(msg, keys.Common.Format):
			cmd := tuitask.CreateTasks("format", resource.GlobalResource, m.ModuleService.Format, m.module.ID)
			return m, cmd
		case key.Matches(msg, keys.Common.Validate):
			cmd := tuitask.CreateTasks("validate", resource.GlobalResource, m.ModuleService.Validate, m.module.ID)
			return m, cmd
		case key.Matches(msg, keys.Common.Edit):
			return m, tui.OpenVim(m.module.Path)
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			return m, tuitask.CreateTasks("reload-workspaces", resource.GlobalResource, m.WorkspaceService.Reload, m.module.ID)
		}
	case resource.Event[*module.Module]:
		if msg.Payload.ID == m.module.ID {
			m.module = msg.Payload
		}
	}
	// Update tabs
	updated, cmd := m.tabs.Update(msg)
	cmds = append(cmds, cmd)
	m.tabs = updated

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return m.helpers.Breadcrumbs("Module", m.module)
}

func (m model) View() string {
	return m.tabs.View()
}

func (m model) HelpBindings() []key.Binding {
	return append(
		m.tabs.HelpBindings(),
		keys.Common.Init,
		keys.Common.Validate,
		keys.Common.Format,
		keys.Common.Plan,
		keys.Common.Edit,
		localKeys.ReloadWorkspaces,
	)
}
