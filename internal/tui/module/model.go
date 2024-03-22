package module

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	runtui "github.com/leg100/pug/internal/tui/run"
	tasktui "github.com/leg100/pug/internal/tui/task"
	workspacetui "github.com/leg100/pug/internal/tui/workspace"
	"github.com/leg100/pug/internal/workspace"
)

const (
	workspacesTabTitle = "workspaces"
	runsTabTitle       = "runs"
	tasksTabTitle      = "tasks"
)

// Maker makes module models.
type Maker struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service

	WorkspaceListMaker *workspacetui.ListMaker
	RunListMaker       *runtui.ListMaker
	TaskListMaker      *tasktui.ListMaker
}

func (mm *Maker) Make(mr resource.Resource, width, height int) (tui.Model, error) {
	mod, err := mm.ModuleService.Get(mr.ID())
	if err != nil {
		return model{}, err
	}

	tabs := tui.NewTabSet(width, height)
	if _, err := tabs.AddTab(mm.WorkspaceListMaker, mr, workspacesTabTitle); err != nil {
		return nil, fmt.Errorf("adding workspaces tab: %w", err)
	}
	if _, err := tabs.AddTab(mm.RunListMaker, mr, runsTabTitle); err != nil {
		return nil, fmt.Errorf("adding runs tab: %w", err)
	}
	if _, err := tabs.AddTab(mm.TaskListMaker, mr, tasksTabTitle); err != nil {
		return nil, fmt.Errorf("adding tasks tab: %w", err)
	}

	m := model{
		ModuleService: mm.ModuleService,
		module:        mod,
		tabs:          tabs,
	}
	return m, nil
}

type model struct {
	ModuleService *module.Service

	module *module.Module
	tabs   tui.TabSet
}

func (m model) Init() tea.Cmd {
	return m.tabs.Init()
}

func (m model) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Init):
			m.tabs.SetActiveTabWithTitle(tasksTabTitle)
			return m, tui.CreateTasks(m.ModuleService.Init, m.module.ID())
		}
	case resource.Event[*module.Module]:
		if msg.Payload.ID() == m.module.ID() {
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
	return tui.Breadcrumbs("Module", m.module.Resource)
}

func (m model) View() string {
	return m.tabs.View()
}

func (m model) Pagination() string {
	return ""
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return keys.KeyMapToSlice(workspacetui.Keys)
}
