package run

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	tasktui "github.com/leg100/pug/internal/tui/task"
)

type Maker struct {
	RunService  tui.RunService
	TaskService tui.TaskService
	Spinner     *spinner.Model
	Helpers     *tui.Helpers
}

func (mm *Maker) Make(rr resource.Resource, width, height int) (tea.Model, error) {
	run, err := mm.RunService.Get(rr.ID)
	if err != nil {
		return model{}, err
	}

	taskMaker := &tasktui.Maker{
		TaskService: mm.TaskService,
		Spinner:     mm.Spinner,
		IsRunTab:    true,
	}

	m := model{
		svc:       mm.RunService,
		tasks:     mm.TaskService,
		run:       run,
		taskMaker: taskMaker,
		helpers:   mm.Helpers,
	}
	m.tabs = tui.NewTabSet(width, height).WithTabSetInfo(&m)

	// Add tabs for existing tasks
	tasks := mm.TaskService.List(task.ListOptions{
		Ancestor: rr.ID,
		// Ensures the plan tab is rendered first
		Oldest: true,
	})
	for _, t := range tasks {
		_, err := m.tabs.AddTab(taskMaker, t.Resource, t.CommandString())
		if err != nil {
			return nil, err
		}
	}
	// If there is an apply task tab, then make that the active tab.
	m.tabs.SetActiveTab("apply")

	return m, nil
}

type model struct {
	svc       tui.RunService
	tasks     tui.TaskService
	run       *run.Run
	tabs      tui.TabSet
	taskMaker tui.Maker
	helpers   *tui.Helpers
}

func (m model) Init() tea.Cmd {
	return m.tabs.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Apply):
			return m, func() tea.Msg {
				if _, err := m.svc.Apply(m.run.ID); err != nil {
					return tui.NewErrorMsg(err, "applying run")
				}
				return nil
			}
		case key.Matches(msg, keys.Common.Module):
			return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(*m.run.Module()))
		case key.Matches(msg, keys.Common.Workspace):
			return m, tui.NavigateTo(tui.WorkspaceKind, tui.WithParent(*m.run.Workspace()))
		}
	case resource.Event[*run.Run]:
		if msg.Payload.ID == m.run.ID {
			m.run = msg.Payload
		}
	case resource.Event[*task.Task]:
		// Create tab for new run task
		switch msg.Type {
		case resource.CreatedEvent:
			if !msg.Payload.HasAncestor(m.run.ID) {
				break
			}
			cmd, err := m.addTab(msg.Payload)
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			// Initialize the new tab
			cmds = append(cmds, cmd)
		}
	}
	// Update tabs
	updated, cmd := m.tabs.Update(msg)
	cmds = append(cmds, cmd)
	m.tabs = updated

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return m.helpers.Breadcrumbs("Run", *m.run.Parent)
}

func (m model) Status() string {
	return m.helpers.RunStatus(m.run)
}

func (m model) ID() string {
	return m.run.String()
}

func (m *model) addTab(t *task.Task) (tea.Cmd, error) {
	title := t.CommandString()
	cmd, err := m.tabs.AddTab(m.taskMaker, t.Resource, title)
	if err != nil {
		// Silently ignore attempts to add duplicate tabs: this can happen when
		// a task is received in both a created event as well as in the initial
		// listing of existing tasks, which is not unlikely.
		if errors.Is(err, tui.ErrDuplicateTab) {
			return nil, nil
		}
		return nil, fmt.Errorf("adding %s tab: %w", title, err)
	}
	// Make the newly added tab the active tab.
	m.tabs.SetActiveTab(title)
	return cmd, nil
}
func (m model) View() string {
	return m.tabs.View()
}

func (m model) TabSetInfo() string {
	hasTabs, activeTab := m.tabs.Active()
	if !hasTabs {
		return ""
	}
	switch activeTab.Title {
	case "plan":
		return m.helpers.RunReport(m.run.PlanReport)
	case "apply":
		return m.helpers.RunReport(m.run.ApplyReport)
	default:
		return ""
	}
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Apply,
	}
}
