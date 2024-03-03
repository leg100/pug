package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
)

type taskListModel struct {
	table table.Model
}

func NewTaskListModel(svc *task.Service, parent *resource.Resource) taskListModel {
	var opts task.ListOptions
	if parent != nil {
		opts.Ancestor = parent.ID
	}

	tasks := svc.List(opts)
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		{Title: "ID", Width: 10},
	}
	columns = append(columns, table.Column{Title: "MODULE", Width: 10})
	columns = append(columns, table.Column{Title: "WORKSPACE", Width: 10})
	columns = append(columns, table.Column{Title: "COMMAND", Width: 10})
	columns = append(columns, table.Column{Title: "STATUS", Width: 10})
	columns = append(columns, table.Column{Title: "AGE", Width: 10})

	rows := make([]table.Row, len(tasks))
	for i, t := range tasks {
		rows[i] = newTaskRow(t)
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	return taskListModel{table: tbl}
}

func newTaskRow(t *task.Task) table.Row {
	return table.Row{
		t.String(),
		t.Module().String(),
		t.Workspace().String(),
		fmt.Sprintf("%v", t.Command),
		string(t.State),
		t.Updated.Round(time.Second).String(),
	}
}

func (m taskListModel) Init() tea.Cmd {
	return nil
}

func (m taskListModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
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
				return m, common.NewErrorCmd(err, "selecting task")
			}
			return m, common.Navigate(common.TaskPage, &resource.Resource{ID: id})
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height)
		return m, nil
	case resource.Event[*task.Task]:
		switch msg.Type {
		case resource.CreatedEvent:
			// Insert new task at top
			m.table.SetRows(
				append([]table.Row{newTaskRow(msg.Payload)}, m.table.Rows()...),
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
				append([]table.Row{newTaskRow(msg.Payload)}, rows...),
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

func (m taskListModel) Title() string {
	return "global tasks"
}

func (m taskListModel) View() string {
	return m.table.View()
}

func (m taskListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

func (m taskListModel) findRow(id resource.ID) int {
	encoded := id.String()
	for i, row := range m.table.Rows() {
		if row[0] == encoded {
			return i
		}
	}
	return -1
}
