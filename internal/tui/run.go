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
	"github.com/leg100/pug/internal/tui/common"
)

type runModelMaker struct {
	svc     *run.Service
	tasks   *task.Service
	spinner *spinner.Model
}

func (mm *runModelMaker) makeModel(rr resource.Resource) (Model, error) {
	run, err := mm.svc.Get(rr.ID())
	if err != nil {
		return runModel{}, err
	}
	m := runModel{
		svc:     mm.svc,
		tasks:   mm.tasks,
		run:     run,
		spinner: mm.spinner,
	}
	return m, nil
}

type runModel struct {
	svc   *run.Service
	tasks *task.Service
	run   *run.Run

	tabs      []tab
	activeTab int

	width  int
	height int

	spinner *spinner.Model
}

type tab struct {
	model Model
	task  *task.Task
}

type initRunTasksMsg []*task.Task

// Init retrieves the run's existing tasks.
func (m runModel) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.tasks.List(task.ListOptions{
			Ancestor: m.run.ID(),
		})
		return initRunTasksMsg(tasks)
	}
}

func (m runModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Apply):
			return m, func() tea.Msg {
				if _, err := m.svc.Apply(m.run.ID()); err != nil {
					return newErrorMsg(err, "applying run")
				}
				return nil
			}
		case key.Matches(msg, Keys.Tab):
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			return m, nil
		case key.Matches(msg, Keys.TabLeft):
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
	case initRunTasksMsg:
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
		m.tabs[i].model, cmd = tm.model.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m runModel) Title() string {
	return breadcrumbs("Run", m.run.Resource)
}

func (m *runModel) createTab(task *task.Task) tea.Cmd {
	// dont create a tab if there is already a tab for the given task.
	for _, existing := range m.tabs {
		if existing.task.ID() == task.ID() {
			return nil
		}
	}
	opts := makeTaskModelOptions{heightAdjustment: 0, isChild: true}
	model, err := makeTaskModel(m.tasks, task.Resource, opts)
	if err != nil {
		return newErrorCmd(err, "creating run task tab")
	}
	tab := tab{
		model: model,
		task:  task,
	}
	m.tabs = append(m.tabs, tab)
	return tab.model.Init()
}

func (m runModel) View() string {
	var (
		tabComponents          []string
		activeTabStyle         = Bold.Copy().Foreground(lipgloss.Color("13"))
		activeStatusStyle      = Regular.Copy().Foreground(lipgloss.Color("13"))
		inactiveTabStyle       = Regular.Copy().Foreground(lipgloss.Color("250"))
		inactiveStatusStyle    = Regular.Copy().Foreground(lipgloss.Color("250"))
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
		underline := headingStyle.Render(strings.Repeat(underlineChar, Width(heading)))
		rendered := lipgloss.JoinVertical(lipgloss.Top, heading, underline)
		tabComponents = append(tabComponents, rendered)
		renderedTabsTotalWidth += Width(heading)
	}
	remainingWidth := max(0, m.width-renderedTabsTotalWidth)
	runStatusSection := lipgloss.JoinVertical(lipgloss.Top,
		// run status
		Regular.Copy().Width(remainingWidth).Align(lipgloss.Right).Padding(0, 1).Render(string(m.run.Status)),
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

func (m runModel) Pagination() string {
	return fmt.Sprintf("tabs: %d; active: %d", len(m.tabs), m.activeTab)
}

func (m runModel) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		Keys.Apply,
	}
}
