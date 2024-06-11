package workspace

import (
	"errors"
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
	"github.com/leg100/pug/internal/tui/split"
	"github.com/leg100/pug/internal/tui/table"
)

var resourceColumn = table.Column{
	Key:        "resource",
	Title:      "RESOURCE",
	FlexFactor: 1,
}

type ResourceListMaker struct {
	StateService tui.StateService
	RunService   tui.RunService
	Spinner      *spinner.Model
	Helpers      *tui.Helpers
}

func (m *ResourceListMaker) Make(ws resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{resourceColumn}
	renderer := func(resource *state.Resource) table.RenderedRow {
		addr := string(resource.Address)
		if resource.Tainted {
			addr += " (tainted)"
		}
		return table.RenderedRow{resourceColumn.Key: addr}
	}
	tableOptions := []table.Option[resource.ID, *state.Resource]{
		table.WithSortFunc(state.Sort),
		table.WithParent[resource.ID, *state.Resource](ws),
	}
	splitModel := split.New(split.Options[*state.Resource]{
		Columns:      columns,
		Renderer:     renderer,
		TableOptions: tableOptions,
		Width:        width,
		Height:       height,
		Maker: &ResourceMaker{
			Helpers:        m.Helpers,
			disableBorders: true,
		},
	})
	return resourceList{
		Model:     splitModel,
		states:    m.StateService,
		runs:      m.RunService,
		workspace: ws,
		spinner:   m.Spinner,
		width:     width,
		height:    height,
		helpers:   m.Helpers,
	}, nil
}

type resourceList struct {
	split.Model[*state.Resource]

	states      tui.StateService
	runs        tui.RunService
	workspace   resource.Resource
	state       *state.State
	reloading   bool
	height      int
	width       int
	helpers     *tui.Helpers
	spinner     *spinner.Model
	initialized bool
}

type initState *state.State

func (m resourceList) Init() tea.Cmd {
	initState := func() tea.Msg {
		if m.initialized {
			return nil
		}
		state, err := m.states.Get(m.workspace.GetID())
		if err != nil {
			return tui.ReportError(err, "initializing state model")
		}
		return initState(state)
	}
	return tea.Batch(initState, m.Model.Init())
}

// reloadedMsg is sent when a state reload has finished.
type reloadedMsg struct {
	workspaceID resource.ID
	err         error
}

func (m resourceList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.Table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.ResourceKind, tui.WithParent(row.Value))
			}
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
				m.helpers.CreateTasks("state-rm", fn, m.workspace.GetID()),
			)
		case key.Matches(msg, resourcesKeys.Taint):
			addrs := m.selectedOrCurrentAddresses()
			return m, m.createStateCommand("taint", m.states.Taint, addrs...)
		case key.Matches(msg, resourcesKeys.Untaint):
			addrs := m.selectedOrCurrentAddresses()
			return m, m.createStateCommand("untaint", m.states.Untaint, addrs...)
		case key.Matches(msg, resourcesKeys.Move):
			if row, ok := m.Table.CurrentRow(); ok {
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
						return m.helpers.CreateTasks("state-mv", fn, m.workspace.GetID())
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
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return m.runs.Plan(workspaceID, createRunOptions)
			}
			return m, m.helpers.CreateTasks("plan", fn, m.workspace.GetID())
		}
	case initState:
		if msg.WorkspaceID != m.workspace.GetID() {
			return m, nil
		}
		m.state = (*state.State)(msg)
		m.Table.SetItems(toTableItems(m.state))
	case resource.Event[*state.State]:
		if msg.Payload.WorkspaceID != m.workspace.GetID() {
			return m, nil
		}
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			// Whenever state is created or updated, re-populate table with
			// resources.
			m.Table.SetItems(toTableItems(msg.Payload))
			m.state = msg.Payload
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Handle keyboard and mouse events in the table widget
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m resourceList) View() string {
	border := tui.Regular.Copy().
		Padding(0, 1).
		Border(lipgloss.NormalBorder()).
		// Subtract 2 to accomodate borders
		Width(m.width - 2).
		// Subtract 2 to accomodate borders
		Height(m.height - 2)

	if m.reloading {
		return border.Render(fmt.Sprintf("Pulling state %s", m.spinner.View()))
	}
	if m.state == nil || m.state.Serial < 0 {
		return border.Render("No state found")
	}
	//metadata := fmt.Sprintf("Serial: %d | Terraform Version: %s | Lineage: %s", m.state.Serial, m.state.TerraformVersion, m.state.Lineage)
	return lipgloss.JoinVertical(lipgloss.Left,
		m.Model.View(),
		//strings.Repeat("─", m.width),
		//tui.Regular.Copy().
		//	Margin(0, 1).
		//	Render(
		//		tui.Regular.Copy().
		//			Inline(true).
		//			Render(metadata),
		//	),
	)
}

func (m resourceList) Title() string {
	title := fmt.Sprintf(
		"%s[%s]",
		m.helpers.Breadcrumbs("State", m.workspace),
		m.Table.TotalString(),
	)
	if m.state != nil {
		title += tui.Regular.Copy().
			Background(tui.LightBlue).
			Foreground(tui.Black).
			Padding(0, 1).
			Render(fmt.Sprintf("serial:%d", m.state.Serial))
	}
	return title
}

func (m resourceList) HelpBindings() []key.Binding {
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

func (m resourceList) selectedOrCurrentAddresses() []state.ResourceAddress {
	rows := m.Table.SelectedOrCurrent()
	addrs := make([]state.ResourceAddress, len(rows))
	var i int
	for _, v := range rows {
		addrs[i] = v.Value.Address
		i++
	}
	return addrs
}

func toTableItems(s *state.State) map[resource.ID]*state.Resource {
	to := make(map[resource.ID]*state.Resource, len(s.Resources))
	for _, v := range s.Resources {
		to[v.ID] = v
	}
	return to
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
