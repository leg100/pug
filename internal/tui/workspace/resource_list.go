package workspace

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/split"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

var resourceColumn = table.Column{
	Key:        "resource",
	Title:      "RESOURCE",
	FlexFactor: 1,
}

type ResourceListMaker struct {
	Workspaces *workspace.Service
	States     *state.Service
	Plans      *plan.Service
	Spinner    *spinner.Model
	Helpers    *tui.Helpers
}

func (m *ResourceListMaker) Make(id resource.ID, width, height int) (tea.Model, error) {
	ws, err := m.Workspaces.Get(id)
	if err != nil {
		return nil, err
	}

	columns := []table.Column{resourceColumn}
	renderer := func(resource *state.Resource) table.RenderedRow {
		addr := string(resource.Address)
		if resource.Tainted {
			addr += " (tainted)"
		}
		return table.RenderedRow{resourceColumn.Key: addr}
	}
	tableOptions := []table.Option[*state.Resource]{
		table.WithSortFunc(state.Sort),
		table.WithParent[*state.Resource](ws),
	}
	splitModel := split.New(split.Options[*state.Resource]{
		Columns:      columns,
		Renderer:     renderer,
		TableOptions: tableOptions,
		Width:        width,
		Height:       height,
		Maker: &ResourceMaker{
			States:         m.States,
			Plans:          m.Plans,
			Helpers:        m.Helpers,
			disableBorders: true,
		},
	})
	return resourceList{
		Model:     splitModel,
		states:    m.States,
		plans:     m.Plans,
		workspace: ws,
		spinner:   m.Spinner,
		width:     width,
		height:    height,
		helpers:   m.Helpers,
	}, nil
}

type resourceList struct {
	split.Model[*state.Resource]

	states    *state.Service
	plans     *plan.Service
	workspace resource.Resource
	state     *state.State
	reloading bool
	height    int
	width     int
	helpers   *tui.Helpers

	spinner *spinner.Model
}

type initState *state.State

func (m resourceList) Init() tea.Cmd {
	return func() tea.Msg {
		state, err := m.states.Get(m.workspace.GetID())
		if err != nil {
			return tui.ReportError(fmt.Errorf("initializing state model: %w", err))
		}
		return initState(state)
	}
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
		createRunOptions plan.CreateOptions
	)

	switch msg := msg.(type) {
	case reloadedMsg:
		m.reloading = false
		if msg.err != nil {
			return m, tui.ReportError(fmt.Errorf("reloading state failed: %w", msg.err))
		}
		return m, tui.ReportInfo("reloading finished")
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.Table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.ResourceKind, tui.WithParent(row.ID))
			}
		case key.Matches(msg, resourcesKeys.Reload):
			if m.reloading {
				return m, tui.ReportError(errors.New("reloading in progress"))
			}
			m.reloading = true
			return m, func() tea.Msg {
				msg := reloadedMsg{workspaceID: m.workspace.GetID()}
				if spec, err := m.states.Reload(msg.workspaceID); err != nil {
					msg.err = err
				} else {
					task, err := m.helpers.Tasks.Create(spec)
					if err != nil {
						msg.err = err
					} else if err := task.Wait(); err != nil {
						msg.err = err
					}
				}
				return msg
			}
		case key.Matches(msg, keys.Common.Delete):
			addrs := m.selectedOrCurrentAddresses()
			if len(addrs) == 0 {
				// no rows; do nothing
				return m, nil
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.states.Delete(workspaceID, addrs...)
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Delete %d resource(s)?", len(addrs)),
				m.helpers.CreateTasks(fn, m.workspace.GetID()),
			)
		case key.Matches(msg, resourcesKeys.Taint):
			addrs := m.selectedOrCurrentAddresses()
			return m, m.createStateCommand(m.states.Taint, addrs...)
		case key.Matches(msg, resourcesKeys.Untaint):
			addrs := m.selectedOrCurrentAddresses()
			return m, m.createStateCommand(m.states.Untaint, addrs...)
		case key.Matches(msg, resourcesKeys.Move):
			if row, ok := m.Table.CurrentRow(); ok {
				from := row.Value.Address
				return m, m.helpers.Move(m.workspace.GetID(), from)
			}
		case key.Matches(msg, keys.Common.PlanDestroy):
			// Create a targeted destroy plan.
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			// Create a targeted plan.
			createRunOptions.TargetAddrs = m.selectedOrCurrentAddresses()
			// NOTE: even if the user hasn't selected any rows, we still proceed
			// to create a run without targeted resources.
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.plans.Plan(workspaceID, createRunOptions)
			}
			return m, m.helpers.CreateTasks(fn, m.workspace.GetID())
		}
	case initState:
		if msg.WorkspaceID != m.workspace.GetID() {
			return m, nil
		}
		m.state = (*state.State)(msg)
		m.Table.SetItems(maps.Values(m.state.Resources)...)
	case resource.Event[*state.State]:
		if msg.Payload.WorkspaceID != m.workspace.GetID() {
			return m, nil
		}
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			// Whenever state is created or updated, re-populate table with
			// resources.
			m.Table.SetItems(maps.Values(msg.Payload.Resources)...)
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
	border := tui.Regular.
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
		//tui.Regular.
		//	Margin(0, 1).
		//	Render(
		//		tui.Regular.
		//			Inline(true).
		//			Render(metadata),
		//	),
	)
}

func (m resourceList) Title() string {
	var serial string
	if m.state != nil {
		serial = serialBreadcrumb(m.state.Serial)
	}
	return tui.Breadcrumbs("State", m.workspace, serial)
}

func (m resourceList) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Plan,
		keys.Common.Destroy,
		keys.Common.Delete,
		resourcesKeys.Move,
		resourcesKeys.Taint,
		resourcesKeys.Untaint,
		resourcesKeys.Reload,
	}
	return append(bindings, keys.KeyMapToSlice(split.Keys)...)
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

func serialBreadcrumb(serial int64) string {
	return tui.TitleSerial.Render(fmt.Sprintf("%d", serial))
}
