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

type resourceListMaker struct {
	StateService *state.Service
	Spinner      *spinner.Model
}

func (m *resourceListMaker) Make(ws resource.Resource, width, height int) (tui.Model, error) {
	renderer := func(resource *state.Resource, inherit lipgloss.Style) table.RenderedRow {
		return table.RenderedRow{
			resourceColumn.Key: resource.String(),
		}
	}
	table := table.New([]table.Column{resourceColumn}, renderer, width, height)
	table = table.WithParent(ws)

	return resources{
		table:     table,
		svc:       m.StateService,
		workspace: ws,
		spinner:   m.Spinner,
	}, nil
}

type resources struct {
	table     table.Model[*state.Resource]
	svc       *state.Service
	workspace resource.Resource

	spinner *spinner.Model

	reloading bool
}

func (m resources) Init() tea.Cmd {
	return m.reload()
}

func (m resources) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Reload):
			// Unload existing items from table (otherwise we'll get duplicates).
			m.table.SetItems(nil)
			// Enable spinner
			m.reloading = true
			return m, m.reload()
		case key.Matches(msg, localKeys.Delete):
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				addrs := m.HighlightedOrSelectedAddresses()
				return m.svc.RemoveItems(workspaceID, addrs...)
			}
			return m, tui.CreateTasks("state-rm", fn, m.workspace.ID())
		}
	case reloadedResourcesMsg:
		if msg.workspaceID != m.workspace.ID() {
			return m, nil
		}
		// Disable reloading spinner
		m.reloading = false
		// Forward items onto the table (it is expecting a table.BulkInsertMsg
		// message).
		m.table, cmd = m.table.Update(msg.items)
		cmds = append(cmds, cmd)
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m resources) reload() tea.Cmd {
	return func() tea.Msg {
		resources, err := m.svc.ListResources(m.workspace.ID())
		if err != nil {
			return tui.NewErrorMsg(err, "initialising resources tab")
		}
		return reloadedResourcesMsg{
			workspaceID: m.workspace.ID(),
			items:       table.BulkInsertMsg[*state.Resource](resources),
		}
	}
}

func (m resources) Title() string {
	return ""
}

func (m resources) View() string {
	if m.reloading {
		return lipgloss.NewStyle().
			Margin(0, 1).
			Render(m.spinner.View() + " refreshing state...")
	}
	return m.table.View()
}

func (m resources) Pagination() string {
	return ""
}

func (m resources) TabStatus() string {
	return fmt.Sprintf("(%d)", len(m.table.Items()))
}

func (m resources) HelpBindings() (bindings []key.Binding) {
	return keys.KeyMapToSlice(localKeys)
}

// HighlightedOrSelectedAddresses returns the resource addresses of any modules
// that are currently selected or highlighted.
func (m resources) HighlightedOrSelectedAddresses() (addrs []string) {
	for _, res := range m.table.HighlightedOrSelected() {
		addrs = append(addrs, res.String())
	}
	return
}

// reloadedResourcesMsg is sent once state resources have been reloaded for a
// workspace.
type reloadedResourcesMsg struct {
	workspaceID resource.ID
	items       table.BulkInsertMsg[*state.Resource]
}
