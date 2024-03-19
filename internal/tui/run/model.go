package tui

import (
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
	"github.com/leg100/pug/internal/tui/common"
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
	m := model{
		svc:     mm.RunService,
		tasks:   mm.TaskService,
		run:     run,
		spinner: mm.Spinner,
		tabMaker: &tasktui.Maker{
			TaskService: mm.TaskService,
			IsRunTab:    true,
		},
	}
	return m, nil
}

type model struct {
	svc   *run.Service
	tasks *task.Service
	run   *run.Run

	tabs      []tab
	activeTab int
	tabMaker  *tasktui.Maker

	width  int
	height int

	spinner *spinner.Model
}

type tab struct {
	model tui.Model
	task  *task.Task
}

type initTasksMsg []*task.Task

// Init retrieves the run's existing tasks.
func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.tasks.List(task.ListOptions{
			Ancestor: m.run.ID(),
		})
		return initTasksMsg(tasks)
	}
}

func (m model) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Apply):
			return m, func() tea.Msg {
				if _, err := m.svc.Apply(m.run.ID()); err != nil {
					return tui.NewErrorMsg(err, "applying run")
				}
				return nil
			}
		case key.Matches(msg, tui.Keys.Tab):
			// Cycle tabs, going back to tab #0 after the last tab
			if m.activeTab == len(m.tabs)-1 {
				m.activeTab = 0
			} else {
				m.activeTab = m.activeTab + 1
			}
			return m, nil
		case key.Matches(msg, tui.Keys.TabLeft):
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
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
			cmds = append(cmds, m.createTab(msg.Payload))
		}
	case initTasksMsg:
		// Create tabs for existing run tasks
		for _, t := range msg {
			if !t.HasAncestor(m.run.ID()) {
				continue
			}
			cmds = append(cmds, m.createTab(t))
		}
		return m, tea.Batch(cmds...)
	case common.ViewSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Relay messages onto child task models
	for i, tm := range m.tabs {
		updated, cmd := tm.model.Update(msg)
		cmds = append(cmds, cmd)
		m.tabs[i].model = updated
	}

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return tui.Breadcrumbs("Run", m.run.Resource)
}

func (m *model) createTab(task *task.Task) tea.Cmd {
	// dont create a tab if there is already a tab for the given task.
	for _, existing := range m.tabs {
		if existing.task.ID() == task.ID() {
			return nil
		}
	}
	model, err := m.tabMaker.Make(task.Resource, m.width, m.height)
	if err != nil {
		return tui.NewErrorCmd(err, "creating run task tab")
	}
	tab := tab{
		model: model,
		task:  task,
	}
	m.tabs = append(m.tabs, tab)
	// switch view to the newly created tab
	m.activeTab = len(m.tabs) - 1
	return tab.model.Init()
}

func (m model) View() string {
	var (
		tabComponents          []string
		activeTabStyle         = tui.Bold.Copy().Foreground(lipgloss.Color("13"))
		activeStatusStyle      = tui.Regular.Copy().Foreground(lipgloss.Color("13"))
		inactiveTabStyle       = tui.Regular.Copy().Foreground(lipgloss.Color("250"))
		inactiveStatusStyle    = tui.Regular.Copy().Foreground(lipgloss.Color("250"))
		renderedTabsTotalWidth int
	)
	for i, t := range m.tabs {
		var (
			headingStyle  lipgloss.Style
			statusStyle   lipgloss.Style
			underlineChar string
		)
		if i == m.activeTab {
			headingStyle = activeTabStyle.Copy()
			statusStyle = activeStatusStyle.Copy()
			underlineChar = "━"
		} else {
			headingStyle = inactiveTabStyle.Copy()
			statusStyle = inactiveStatusStyle.Copy()
			underlineChar = "─"
		}
		heading := headingStyle.Copy().Padding(0, 1).Render(strings.Join(t.task.Command, " "))
		var statusSymbol string
		switch t.task.State {
		case task.Running:
			statusSymbol = m.spinner.View()
		case task.Exited:
			statusSymbol = "✓"
		case task.Errored:
			statusSymbol = "✗"
		}
		heading += statusStyle.Padding(0, 1, 0, 0).Render(statusSymbol)
		underline := headingStyle.Render(strings.Repeat(underlineChar, tui.Width(heading)))
		rendered := lipgloss.JoinVertical(lipgloss.Top, heading, underline)
		tabComponents = append(tabComponents, rendered)
		renderedTabsTotalWidth += tui.Width(heading)
	}
	remainingWidth := max(0, m.width-renderedTabsTotalWidth)
	runStatusSection := lipgloss.JoinVertical(lipgloss.Top,
		// run status
		tui.Regular.Copy().Width(remainingWidth).Align(lipgloss.Right).Padding(0, 1).Render(string(m.run.Status)),
		// faint grey underline
		inactiveTabStyle.Copy().Render(strings.Repeat("─", remainingWidth)),
	)
	tabComponents = append(tabComponents, runStatusSection)

	tabSection := lipgloss.JoinHorizontal(lipgloss.Bottom, tabComponents...)
	var tabContent string
	if len(m.tabs) > 0 {
		tabContent = m.tabs[m.activeTab].model.View()
	}
	return lipgloss.JoinVertical(lipgloss.Top, tabSection, tabContent)
}

func (m model) Pagination() string {
	return fmt.Sprintf("tabs: %d; active: %d", len(m.tabs), m.activeTab)
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		tui.Keys.Apply,
	}
}
