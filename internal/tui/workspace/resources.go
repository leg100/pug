package workspace

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
	FlexFactor: 1,
}

type resourceListMaker struct {
	StateService tui.StateService
	RunService   tui.RunService
	Spinner      *spinner.Model
}

func (m *resourceListMaker) Make(ws resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{resourceColumn}
	renderer := func(resource *state.Resource) table.RenderedRow {
		addr := string(resource.Address)
		if resource.Tainted {
			addr += " (tainted)"
		}
		return table.RenderedRow{resourceColumn.Key: addr}
	}
	table := table.New(columns, renderer, width, height-metadataHeight).
		WithSortFunc(state.Sort).
		WithParent(ws)
	return resources{
		table:     table,
		states:    m.StateService,
		runs:      m.RunService,
		workspace: ws,
		spinner:   m.Spinner,
		width:     width,
	}, nil
}

type resources struct {
	table     table.Model[resource.ID, *state.Resource]
	states    tui.StateService
	runs      tui.RunService
	workspace resource.Resource
	state     *state.State
	reloading bool
	width     int

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

// reloadedMsg is sent when a state reload has finished.
type reloadedMsg struct {
	workspaceID resource.ID
	err         error
}

func (m resources) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions run.CreateOptions
	)

	switch msg := msg.(type) {
	case reloadedMsg:
		m.reloading = false
		if msg.err != nil {
			return m, tui.ReportError(msg.err, "reloading state failed")
		}
		return m, tui.ReportInfo("reloading finished")
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, resourcesKeys.Reload):
			if m.reloading {
				return m, tui.ReportError(errors.New("reloading in progress"), "")
			}
			m.reloading = true
			return m, func() tea.Msg {
				msg := reloadedMsg{workspaceID: m.workspace.GetID()}
				if task, err := m.states.Reload(msg.workspaceID); err != nil {
					msg.err = err
				} else if err := task.Wait(); err != nil {
					msg.err = err
				}
				return msg
			}
		case key.Matches(msg, keys.Common.Delete):
			addrs := m.selectedOrCurrentAddresses()
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
			addrs := m.selectedOrCurrentAddresses()
			return m, m.createStateCommand("taint", m.states.Taint, addrs...)
		case key.Matches(msg, resourcesKeys.Untaint):
			addrs := m.selectedOrCurrentAddresses()
			return m, m.createStateCommand("untaint", m.states.Untaint, addrs...)
		case key.Matches(msg, resourcesKeys.Move):
			if row, ok := m.table.CurrentRow(); ok {
				from := row.Value.Address
				return m, tui.CmdHandler(tui.PromptMsg{
					Prompt:       "Enter destination address: ",
					InitialValue: string(from),
					Action: func(v string) tea.Cmd {
						if v == "" {
							return nil
						}
						fn := func(workspaceID resource.ID) (*task.Task, error) {
							return m.states.Move(workspaceID, from, state.ResourceAddress(v))
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
			createRunOptions.TargetAddrs = m.selectedOrCurrentAddresses()
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
		m.table.SetItems(toTableItems(m.state))
	case resource.Event[*state.State]:
		if msg.Payload.WorkspaceID != m.workspace.GetID() {
			return m, nil
		}
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			// Whenever state is created or updated, re-populate table with
			// resources.
			m.table.SetItems(toTableItems(msg.Payload))
			m.state = msg.Payload
		}
	}

	wsm, ok := msg.(tea.WindowSizeMsg)
	if ok {
		m.width = wsm.Width
		// adjust height to accomodate metadata section before the message is
		// relayed to the table model.
		wsm.Height -= metadataHeight
		m.table, cmd = m.table.Update(wsm)
	} else {
		// Handle keyboard and mouse events in the table widget
		m.table, cmd = m.table.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

const (
	// metadataHeight is the height of the metadata section beneath the table,
	// including the horizontal rule divider.
	metadataHeight = 2
)

func (m resources) View() string {
	if m.state == nil || m.state.Serial < 0 {
		return tui.Regular.Copy().
			Margin(0, 1).
			Render("No state found")
	}
	metadata := fmt.Sprintf("Serial: %d | Terraform Version: %s | Lineage: %s", m.state.Serial, m.state.TerraformVersion, m.state.Lineage)
	return lipgloss.JoinVertical(lipgloss.Left,
		m.table.View(),
		strings.Repeat("─", m.width),
		tui.Regular.Copy().
			Margin(0, 1).
			Render(
				tui.Regular.Copy().
					Inline(true).
					Render(metadata),
			),
	)
}

func (m resources) TabStatus() string {
	if m.reloading {
		return m.spinner.View()
	}
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

func (m resources) selectedOrCurrentAddresses() []state.ResourceAddress {
	rows := m.table.SelectedOrCurrent()
	addrs := make([]state.ResourceAddress, len(rows))
	var i int
	for _, v := range rows {
		addrs[i] = v.Value.Address
		i++
	}
	return addrs
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

func toTableItems(s *state.State) map[resource.ID]*state.Resource {
	to := make(map[resource.ID]*state.Resource, len(s.Resources))
	for _, v := range s.Resources {
		to[v.ID] = v
	}
	return to
}
