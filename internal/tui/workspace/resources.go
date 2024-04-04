package workspace

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
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
	Spinner      *spinner.Model
}

func (m *resourceListMaker) Make(ws resource.Resource, width, height int) (tui.Model, error) {
	columns := []table.Column{
		resourceColumn,
		resourceStatusColumn,
	}
	renderer := func(resource *state.Resource, inherit lipgloss.Style) table.RenderedRow {
		return table.RenderedRow{
			resourceColumn.Key:       resource.Address.String(),
			resourceStatusColumn.Key: string(resource.Status),
		}
	}
	table := table.New[state.ResourceAddress](columns, renderer, width, height)
	table = table.WithSortFunc(state.Sort)

	return resources{
		table:     table,
		svc:       m.StateService,
		workspace: ws,
		spinner:   m.Spinner,
	}, nil
}

type resources struct {
	table     table.Model[state.ResourceAddress, *state.Resource]
	svc       tui.StateService
	workspace resource.Resource
	state     *state.State

	spinner *spinner.Model
}

type initState *state.State

func (m resources) Init() tea.Cmd {
	return func() tea.Msg {
		state, err := m.svc.Get(m.workspace.ID)
		if err != nil {
			return tui.ReportError(err, "loading resources tab")
		}
		return initState(state)
	}
}

func (m resources) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, resourcesKeys.Reload):
			return m, tui.CreateTasks("reload state", m.svc.Reload, m.workspace.ID)
		case key.Matches(msg, resourcesKeys.Delete):
			addrs := m.table.HighlightedOrSelectedKeys()
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.svc.Delete(workspaceID, addrs...)
			}
			return m, tui.CreateTasks("state-rm", fn, m.workspace.ID)
		case key.Matches(msg, resourcesKeys.Taint):
			addrs := m.table.HighlightedOrSelectedKeys()
			return m, func() tea.Msg {
				tasks, errs := m.taintMany(m.workspace.ID, addrs...)
				return tui.CreatedTasksMsg{
					Command:    "state-taint",
					Tasks:      tasks,
					CreateErrs: errs,
				}
			}
		}
	case initState:
		m.state = (*state.State)(msg)
		m.table.SetItems(m.state.Resources)
	case resource.Event[*state.State]:
		if msg.Payload.WorkspaceID != m.workspace.ID {
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
	return fmt.Sprintf("(%d)", len(m.table.Items()))
}

func (m resources) HelpBindings() (bindings []key.Binding) {
	return keys.KeyMapToSlice(resourcesKeys)
}

func (m resources) taintMany(workspaceID resource.ID, addrs ...state.ResourceAddress) (multi task.Multi, errs []error) {
	for _, addr := range addrs {
		task, err := m.svc.Taint(workspaceID, addr)
		if err != nil {
			errs = append(errs, err)
		}
		multi = append(multi, task)
	}
	return
}

type resourcesKeyMap struct {
	Delete key.Binding
	Taint  key.Binding
	Reload key.Binding
}

var resourcesKeys = resourcesKeyMap{
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Taint: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "taint"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
}
