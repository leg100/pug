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
	States     *state.Service
	Workspaces *workspace.Service
	Plans      *plan.Service
	Spinner    *spinner.Model
	Helpers    *tui.Helpers
}

func (m *ResourceListMaker) Make(workspaceID resource.ID, width, height int) (tui.ChildModel, error) {
	ws, err := m.Workspaces.Get(workspaceID)
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
	tbl := table.New(
		columns,
		renderer,
		width,
		height,
		table.WithSortFunc(state.Sort),
		table.WithPreview[*state.Resource](tui.ResourceKind),
	)
	return &resourceList{
		Model:     tbl,
		states:    m.States,
		plans:     m.Plans,
		workspace: ws,
		spinner:   m.Spinner,
		width:     width,
		height:    height,
		Helpers:   m.Helpers,
	}, nil
}

type resourceList struct {
	table.Model[*state.Resource]
	*tui.Helpers

	states    *state.Service
	plans     *plan.Service
	state     *state.State
	workspace *workspace.Workspace
	reloading bool
	height    int
	width     int

	spinner *spinner.Model
}

type initState *state.State

func (m *resourceList) Init() tea.Cmd {
	return func() tea.Msg {
		state, err := m.states.Get(m.workspace.ID)
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

func (m *resourceList) Update(msg tea.Msg) tea.Cmd {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions plan.CreateOptions
		applyPrompt      = "Auto-apply %d resources?"
	)

	switch msg := msg.(type) {
	case reloadedMsg:
		m.reloading = false
		if msg.err != nil {
			return tui.ReportError(fmt.Errorf("reloading state failed: %w", msg.err))
		}
		return tui.ReportInfo("reloading finished")
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Enter):
			if row, ok := m.CurrentRow(); ok {
				return tui.NavigateTo(tui.ResourceKind, tui.WithParent(row.ID))
			}
		case key.Matches(msg, resourcesKeys.Reload):
			if m.reloading {
				return tui.ReportError(errors.New("reloading in progress"))
			}
			m.reloading = true
			return func() tea.Msg {
				msg := reloadedMsg{workspaceID: m.workspace.ID}
				if spec, err := m.states.Reload(msg.workspaceID); err != nil {
					msg.err = err
				} else {
					task, err := m.Tasks.Create(spec)
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
				return nil
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.states.Delete(workspaceID, addrs...)
			}
			return tui.YesNoPrompt(
				fmt.Sprintf("Delete %d resource(s)?", len(addrs)),
				m.CreateTasks(fn, m.workspace.ID),
			)
		case key.Matches(msg, resourcesKeys.Taint):
			addrs := m.selectedOrCurrentAddresses()
			return m.createStateCommand(m.states.Taint, addrs...)
		case key.Matches(msg, resourcesKeys.Untaint):
			addrs := m.selectedOrCurrentAddresses()
			return m.createStateCommand(m.states.Untaint, addrs...)
		case key.Matches(msg, resourcesKeys.Move):
			if row, ok := m.CurrentRow(); ok {
				from := row.Value.Address
				return m.Move(m.workspace.ID, from)
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
			return m.CreateTasks(fn, m.workspace.ID)
		case key.Matches(msg, keys.Common.Destroy):
			createRunOptions.Destroy = true
			applyPrompt = "Destroy %d resources?"
			fallthrough
		case key.Matches(msg, keys.Common.Apply):
			// Create a targeted apply.
			createRunOptions.TargetAddrs = m.selectedOrCurrentAddresses()
			resourceIDs := m.SelectedOrCurrentIDs()
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.plans.Apply(workspaceID, createRunOptions)
			}
			return tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(resourceIDs)),
				m.CreateTasks(fn, m.workspace.ID),
			)
		}
	case initState:
		if msg.WorkspaceID != m.workspace.ID {
			return nil
		}
		m.state = (*state.State)(msg)
		m.SetItems(maps.Values(m.state.Resources)...)
	case resource.Event[*state.State]:
		if msg.Payload.WorkspaceID != m.workspace.ID {
			return nil
		}
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			// Whenever state is created or updated, re-populate table with
			// resources.
			m.SetItems(maps.Values(msg.Payload.Resources)...)
			m.state = msg.Payload
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Handle keyboard and mouse events in the table widget
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m resourceList) View() string {
	if m.reloading {
		return fmt.Sprintf("Pulling state %s", m.spinner.View())
	}
	if m.state == nil || m.state.Serial < 0 {
		return "No state found"
	}
	// Make footer
	// metadata := fmt.Sprintf("Serial: %d | Terraform Version: %s | Lineage: %s", m.state.Serial, m.state.TerraformVersion, m.state.Lineage)
	return lipgloss.JoinVertical(lipgloss.Left,
		m.Model.View(),
		// strings.Repeat("─", m.width),
		// tui.Regular.
		//	Margin(0, 1).
		//	Render(
		//		tui.Regular.
		//			Inline(true).
		//			Render(metadata),
		//	),
	)
}

func (m resourceList) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.Apply,
		keys.Common.Destroy,
		keys.Common.Delete,
		resourcesKeys.Move,
		resourcesKeys.Taint,
		resourcesKeys.Untaint,
		resourcesKeys.Reload,
	}
}

func (m resourceList) selectedOrCurrentAddresses() []state.ResourceAddress {
	rows := m.SelectedOrCurrent()
	addrs := make([]state.ResourceAddress, len(rows))
	var i int
	for _, v := range rows {
		addrs[i] = v.Value.Address
		i++
	}
	return addrs
}

func (m *resourceList) BorderText() map[tui.BorderPosition]string {
	var serial int64
	if m.state != nil {
		serial = m.state.Serial
	}
	return map[tui.BorderPosition]string{
		tui.TopLeft: fmt.Sprintf(
			"%s %s %s",
			tui.Bold.Render("state"),
			tui.ModulePathWithIcon(m.workspace.ModulePath, true),
			tui.WorkspaceNameWithIcon(m.workspace.Name, true),
		),
		tui.TopMiddle: m.Metadata(),
		tui.BottomMiddle: lipgloss.NewStyle().
			Foreground(tui.BurntOrange).
			Render(fmt.Sprintf("#%d", serial)),
	}
}

func serialBreadcrumb(serial int64) string {
	return tui.TitleSerial.Render(fmt.Sprintf("%d", serial))
}
