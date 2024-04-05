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
		ModuleService: mm.ModuleService,
		RunService:    mm.RunService,
		module:        mod,
		tabs:          tabs,
		helpers:       mm.Helpers,
	}
	return m, nil
}

type model struct {
	ModuleService tui.ModuleService
	RunService    tui.RunService

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
		case key.Matches(msg, localKeys.Init):
			// 'i' creates a terraform init task and sends the user to the tasks
			// tab.
			m.tabs.SetActiveTab(tasksTabTitle)
			return m, tui.CreateTasks("init", m.ModuleService.Init, m.module.ID)
		case key.Matches(msg, localKeys.Edit):
			return m, tui.OpenVim(m.module.Path)
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

	// Tab-specific actions to be taken after action has been sent to tab.
	switch m.tabs.ActiveTitle() {
	case workspacesTabTitle:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, localKeys.Plan):
				// The workspaces tab model takes care of listening to this key
				// press and creating the actual run, and only once that's done
				// do we then send the user to the runs tab.
				m.tabs.SetActiveTab(runsTabTitle)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return m.helpers.Breadcrumbs("Module", m.module.Resource)
}

func (m model) View() string {
	return m.tabs.View()
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return keys.KeyMapToSlice(localKeys)
}
