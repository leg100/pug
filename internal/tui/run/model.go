package run

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	tasktui "github.com/leg100/pug/internal/tui/task"
)

type Maker struct {
	RunService  *run.Service
	TaskService *task.Service
	Spinner     *spinner.Model
}

func (mm *Maker) Make(rr resource.Resource, width, height int) (tui.Model, error) {
	run, err := mm.RunService.Get(rr.ID())
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
	}
	m.tabs = tui.NewTabSet(width, height).WithTabSetInfo(&m)

	// Add tabs for existing tasks
	tasks := mm.TaskService.List(task.ListOptions{
		Ancestor: rr.ID(),
		// Ensures the plan tab is rendered first
		Oldest: true,
	})
	for _, t := range tasks {
		if _, err := m.addTab(t); err != nil {
			return nil, err
		}
	}

	return m, nil
}

type model struct {
	svc       *run.Service
	tasks     *task.Service
	run       *run.Run
	tabs      tui.TabSet
	taskMaker tui.Maker
}

// Init retrieves the run's existing tasks.
func (m model) Init() tea.Cmd {
	return m.tabs.Init()
}

func (m model) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Apply):
			return m, func() tea.Msg {
				if _, err := m.svc.Apply(m.run.ID()); err != nil {
					return tui.NewErrorMsg(err, "applying run")
				}
				return nil
			}
		}
	case resource.Event[*run.Run]:
		if msg.Payload.ID() == m.run.ID() {
			m.run = msg.Payload
		}
	case resource.Event[*task.Task]:
		// Create tab for new run task
		switch msg.Type {
		case resource.CreatedEvent:
			if !msg.Payload.HasAncestor(m.run.ID()) {
				break
			}
			cmd, err := m.addTab(msg.Payload)
			if err != nil {
				return m, tui.ReportError(err, "")
			}
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
	return tui.Breadcrumbs("Run", m.run.Resource)
}

func (m model) Status() string {
	return tui.RenderRunStatus(m.run.Status)
}

func (m model) ID() string {
	return m.run.String()
}

func (m *model) addTab(t *task.Task) (tea.Cmd, error) {
	title := strings.Join(t.Command, " ")
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
	m.tabs.SetActiveTab(-1)
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
		return tui.RenderRunReport(m.run.PlanReport, lipgloss.Style{})
	case "apply":
		return tui.RenderRunReport(m.run.ApplyReport, lipgloss.Style{})
	default:
		return ""
	}
}

func (m model) Pagination() string {
	return ""
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Apply,
	}
}
