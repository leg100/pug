package workspace

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	tuitask "github.com/leg100/pug/internal/tui/task"
)

var resourceColumn = table.Column{
	Key:        "resource",
	Title:      "RESOURCE",
	FlexFactor: 2,
}

var resourceStatusColumn = table.Column{
	Key:        "resource_status",
	Title:      "STATUS",
	FlexFactor: 1,
}

type resourceListMaker struct {
	StateService tui.StateService
	RunService   tui.RunService
	Spinner      *spinner.Model
}

func (m *resourceListMaker) Make(ws resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{
		resourceColumn,
		resourceStatusColumn,
	}
	renderer := func(resource *state.Resource) table.RenderedRow {
		return table.RenderedRow{
			resourceColumn.Key:       string(resource.Address),
			resourceStatusColumn.Key: string(resource.Status),
		}
	}
	table := table.New[state.ResourceAddress](columns, renderer, width, height)
	table = table.WithSortFunc(state.Sort)

	return resources{
		table:     table,
		states:    m.StateService,
		runs:      m.RunService,
		workspace: ws,
		spinner:   m.Spinner,
	}, nil
}

type resources struct {
	table     table.Model[state.ResourceAddress, *state.Resource]
	states    tui.StateService
	runs      tui.RunService
	workspace resource.Resource
	state     *state.State

	spinner *spinner.Model
}

type initState *state.State

func (m resources) Init() tea.Cmd {
	return func() tea.Msg {
		state, err := m.states.Get(m.workspace.GetID())
		if err != nil {
			return tui.ReportError(err, "loading resources tab")
		}
		return initState(state)
	}
}

func (m resources) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, resourcesKeys.Reload):
			return m, tuitask.CreateTasks(tuitask.ReloadStateCommand, m.workspace, m.states.Reload, m.workspace.GetID())
		case key.Matches(msg, keys.Common.Delete):
			addrs := m.table.HighlightedOrSelectedKeys()
			if len(addrs) == 0 {
				// no rows; do nothing
				return m, nil
			}
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.states.Delete(workspaceID, addrs...)
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Delete %d resource(s)?", len(addrs)),
				tuitask.CreateTasks("state-rm", m.workspace, fn, m.workspace.GetID()),
			)
		case key.Matches(msg, resourcesKeys.Taint):
			addrs := m.table.HighlightedOrSelectedKeys()
			return m, m.createStateCommand("taint", m.states.Taint, addrs...)
		case key.Matches(msg, resourcesKeys.Untaint):
			addrs := m.table.HighlightedOrSelectedKeys()
			return m, m.createStateCommand("untaint", m.states.Untaint, addrs...)
		case key.Matches(msg, resourcesKeys.Move):
			if row, highlighted := m.table.Highlighted(); highlighted {
				return m, tui.CmdHandler(tui.PromptMsg{
					Prompt:       "Enter destination address: ",
					InitialValue: string(row.Key),
					Action: func(v string) tea.Cmd {
						if v == "" {
							return nil
						}
						fn := func(workspaceID resource.ID) (*task.Task, error) {
							return m.states.Move(workspaceID, row.Key, state.ResourceAddress(v))
						}
						return tuitask.CreateTasks("state-mv", m.workspace, fn, m.workspace.GetID())
					},
					Key:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
					Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
				})
			}
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			// Create a targeted run.
			createRunOptions.TargetAddrs = m.table.HighlightedOrSelectedKeys()
			// NOTE: even if the user hasn't selected any rows, we still proceed
			// to create a run without targeted resources.
			return m, func() tea.Msg {
				run, err := m.runs.Create(m.workspace.GetID(), createRunOptions)
				if err != nil {
					return tui.NewErrorMsg(err, "creating targeted run")
				}
				return tui.NewNavigationMsg(tui.RunKind, tui.WithParent(run))
			}
		}
	case initState:
		m.state = (*state.State)(msg)
		m.table.SetItems(m.state.Resources)
	case resource.Event[*state.State]:
		if msg.Payload.WorkspaceID != m.workspace.GetID() {
			return m, nil
		}
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			// Whenever state is created or updated, re-populate table with
			// resources.
			m.table.SetItems(msg.Payload.Resources)
		}
		m.state = msg.Payload
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Not used
func (m resources) Title() string {
	return ""
}

func (m resources) View() string {
	if m.state != nil && m.state.State == state.ReloadingState {
		return lipgloss.NewStyle().
			Margin(0, 1).
			Render(m.spinner.View() + " refreshing state...")
	}
	return m.table.View()
}

func (m resources) TabStatus() string {
	return fmt.Sprintf("(%s)", m.table.TotalString())
}

func (m resources) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Plan,
		keys.Common.Destroy,
		keys.Common.Delete,
		resourcesKeys.Move,
		resourcesKeys.Taint,
		resourcesKeys.Untaint,
		resourcesKeys.Reload,
	}
}

type stateFunc func(workspaceID resource.ID, addr state.ResourceAddress) (*task.Task, error)

func (m resources) createStateCommand(name string, fn stateFunc, addrs ...state.ResourceAddress) tea.Cmd {
	return func() tea.Msg {
		msg := tuitask.CreatedTasksMsg{Command: name, Issuer: m.workspace}
		for _, addr := range addrs {
			task, err := fn(m.workspace.GetID(), addr)
			if err != nil {
				msg.CreateErrs = append(msg.CreateErrs, err)
			}
			msg.Tasks = append(msg.Tasks, task)
		}
		return msg
	}
}

type resourcesKeyMap struct {
	Taint   key.Binding
	Untaint key.Binding
	Move    key.Binding
	Reload  key.Binding
}

var resourcesKeys = resourcesKeyMap{
	Taint: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "taint"),
	),
	Untaint: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "untaint"),
	),
	Move: key.NewBinding(
		key.WithKeys("alt+m"),
		key.WithHelp("alt+m", "move"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
}
