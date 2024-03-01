package task

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/resource"
	taskpkg "github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
)

type listModel struct {
	table table.Model
}

func NewListModel(svc *taskpkg.Service, parent uuid.UUID) listModel {
	tasks := svc.List(taskpkg.ListOptions{
		Ancestor: parent,
	})
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		{Title: "MODULE", Width: 10},
		{Title: "WORKSPACE", Width: 10},
		{Title: "COMMAND", Width: 10},
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
	return listModel{table: tbl}
}

func newRow(t *taskpkg.Task) table.Row {
	return table.Row{
		t.Module().String(),
		t.Workspace().String(),
		fmt.Sprintf("%v", t.Command),
		string(t.State),
		t.Updated.Round(time.Second).String(),
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height)
		return m, nil
	case resource.Event[*taskpkg.Task]:
		switch msg.Type {
		case resource.CreatedEvent:
			// Insert new task at top
			m.table.SetRows(
				append([]table.Row{newRow(msg.Payload)}, m.table.Rows()...),
			)
		case resource.UpdatedEvent:
			// TODO: on update event, lookup matching row, update it, move it to the
			// top, and refresh table.
		case resource.DeletedEvent:
			// TODO: on delete event, lookup matching row, delete it, and
			// refresh table.
		}
		return m, nil
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m listModel) Title() string {
	return "global tasks"
}

func (m listModel) View() string {
	return m.table.View()
}
