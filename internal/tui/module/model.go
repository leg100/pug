package module

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	runtui "github.com/leg100/pug/internal/tui/run"
	tasktui "github.com/leg100/pug/internal/tui/task"
	workspacetui "github.com/leg100/pug/internal/tui/workspace"
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

	WorkspaceListMaker *workspacetui.ListMaker
	RunListMaker       *runtui.ListMaker
	TaskListMaker      *tasktui.ListMaker

	Helpers *tui.Helpers
}

func (mm *Maker) Make(mr resource.Resource, width, height int) (tea.Model, error) {
	mod, err := mm.ModuleService.Get(mr.ID)
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
		if m.tabs.ActiveTabFilterFocused() {
			break
		}
		switch {
		case key.Matches(msg, keys.Common.Init):
			// 'i' creates a terraform init task and sends the user to the tasks
			// tab.
			m.tabs.SetActiveTab(tasksTabTitle)
			return m, tui.CreateTasks("init", m.ModuleService.Init, m.module.ID)
		case key.Matches(msg, keys.Common.Edit):
			return m, tui.OpenVim(m.module.Path)
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			return m, tui.CreateTasks("reload-workspaces", m.WorkspaceService.Reload, m.module.ID)
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
	return m.helpers.Breadcrumbs("Module", m.module.Resource)
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
