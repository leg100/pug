package run

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui/common"
)

type runListModel struct {
	table table.Model
}

func NewRunListModel(svc *run.Service, parent *resource.Resource) runListModel {
	var opts run.ListOptions
	if parent != nil {
		opts.ParentID = &parent.ID
	}

	tasks := svc.List(opts)
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "MODULE", Width: 10},
		{Title: "WORKSPACE", Width: 10},
		{Title: "STATUS", Width: 10},
		{Title: "AGE", Width: 10},
	}
	rows := make([]table.Row, len(tasks))
	for i, t := range tasks {
		rows[i] = newRow(t)
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	return runListModel{table: tbl}
}

func newRow(run *run.Run) table.Row {
	return table.Row{
		run.ID.String(),
		run.Module().String(),
		run.Workspace().String(),
		string(run.Status),
		run.Created.Round(time.Second).String(),
	}
}

func (m runListModel) Init() tea.Cmd {
	return nil
}

func (m runListModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Enter):
			// When user presses enter on a run, then it should navigate to the
			// task for its plan if only a plan has been run, or to the task for
			// its apply, if that has been run.
			row := m.table.SelectedRow()
			id, err := resource.IDFromString(row[0])
			if err != nil {
				return m, common.NewErrorCmd(err, "selecting run")
			}
			return m, common.Navigate(common.TaskPage, id)
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height)
		return m, nil
	case resource.Event[*run.Run]:
		switch msg.Type {
		case resource.CreatedEvent:
			// Insert new at top
			m.table.SetRows(
				append([]table.Row{newRow(msg.Payload)}, m.table.Rows()...),
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
				append([]table.Row{newRow(msg.Payload)}, rows...),
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

func (m runListModel) Title() string {
	return "runs"
}

func (m runListModel) View() string {
	return m.table.View()
}

func (m runListModel) findRow(id resource.ID) int {
	encoded := id.String()
	for i, row := range m.table.Rows() {
		if row[0] == encoded {
			return i
		}
	}
	return -1
}
