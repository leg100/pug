package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/common"
	"github.com/leg100/pug/internal/workspace"
)

type workspaceListModel struct {
	table table.Model
}

func NewWorkspaceListModel(svc *workspace.Service, module *resource.Resource) workspaceListModel {
	var opts workspace.ListOptions
	if module != nil {
		opts.ModuleID = &module.ID
	}

	tasks := svc.List(opts)
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "MODULE", Width: 10},
		{Title: "NAME", Width: 10},
	}
	rows := make([]table.Row, len(tasks))
	for i, t := range tasks {
		rows[i] = newWorkspaceRow(t)
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	return workspaceListModel{table: tbl}
}

func newWorkspaceRow(ws *workspace.Workspace) table.Row {
	return table.Row{
		ws.ID.String(),
		ws.Module().String(),
		ws.String(),
	}
}

func (m workspaceListModel) Init() tea.Cmd {
	return nil
}

func (m workspaceListModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Enter):
			row := m.table.SelectedRow()
			id, err := resource.IDFromString(row[0])
			if err != nil {
				return m, common.NewErrorCmd(err, "selecting workspace")
			}
			return m, common.Navigate(common.TaskPage, id)
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height)
		return m, nil
	case resource.Event[*workspace.Workspace]:
		switch msg.Type {
		case resource.CreatedEvent:
			// Insert new workspace at top
			m.table.SetRows(
				append([]table.Row{newWorkspaceRow(msg.Payload)}, m.table.Rows()...),
			)
		case resource.UpdatedEvent:
			i := m.findRow(msg.Payload.ID)
			if i < 0 {
				// TODO: log error
				return m, nil
			}
			// remove row
			rows := append(m.table.Rows()[:i], m.table.Rows()[i+1:]...)
			// add to top
			m.table.SetRows(
				append([]table.Row{newWorkspaceRow(msg.Payload)}, rows...),
			)
		case resource.DeletedEvent:
			i := m.findRow(msg.Payload.ID)
			if i < 0 {
				// TODO: log error
				return m, nil
			}
			// remove row
			m.table.SetRows(
				append(m.table.Rows()[:i], m.table.Rows()[i+1:]...),
			)
		}
		return m, nil
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m workspaceListModel) Title() string {
	return "global tasks"
}

func (m workspaceListModel) View() string {
	return m.table.View()
}

func (m workspaceListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

func (m workspaceListModel) findRow(id resource.ID) int {
	encoded := id.String()
	for i, row := range m.table.Rows() {
		if row[0] == encoded {
			return i
		}
	}
	return -1
}
