package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/evertras/bubble-table/table"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type taskListModelMaker struct {
	svc      *task.Service
	maxTasks int
}

func (m *taskListModelMaker) makeModel(parent resource.Resource) (Model, error) {
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		table.NewColumn(ColKeyID, "ID", resource.IDEncodedMaxLen).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewFlexColumn(ColKeyModule, "MODULE", 2).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyWorkspace, "WORKSPACE", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewFlexColumn(ColKeyCommand, "COMMAND", 1).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyStatus, "STATUS", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyAgo, "AGE", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}

	rowMaker := func(task *task.Task) table.RowData {
		// TODO: categorise and sort tasks according to a specific schema:
		// 1. running (newest first)
		// 2. queued (oldest first)
		// 3. pending (oldest first)
		// 4. finished (newest first)
		row := table.RowData{
			ColKeyID:      task.ID().String(),
			ColKeyCommand: strings.Join(task.Command, " "),
			ColKeyStatus:  string(task.State),
			ColKeyAgo:     ago(time.Now(), task.Updated),
			ColKeyData:    task,
			ColKeyTime:    task.Updated,
		}
		if mod := task.Module(); mod != nil {
			row[ColKeyModule] = mod.String()
		}
		if ws := task.Workspace(); ws != nil {
			row[ColKeyWorkspace] = ws.String()
		}
		return row
	}
	return taskListModel{
		table: newTableModel(tableModelOptions[*task.Task]{
			rowMaker: rowMaker,
			columns:  columns,
		}),
		svc:    m.svc,
		parent: parent,
		max:    m.maxTasks,
	}, nil
}

type taskListModel struct {
	table  tableModel[*task.Task]
	svc    *task.Service
	parent resource.Resource
	max    int
}

func (m taskListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts task.ListOptions
		if m.parent != resource.NilResource {
			opts.Ancestor = m.parent.ID()
		}
		return bulkInsertMsg[*task.Task](m.svc.List(opts))
	}
}

func (m taskListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Enter):
			if ok, task := m.table.highlighted(); ok {
				return m, navigate(page{kind: TaskKind, resource: task.Resource})
			}
		case key.Matches(msg, Keys.Cancel):
			return m, taskCmd(m.svc.Cancel, m.table.highlightedOrSelectedIDs()...)
		}
	}
	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m taskListModel) Title() string {
	return lipgloss.NewStyle().
		Inherit(Breadcrumbs).
		Padding(0, 0, 0, 1).
		Render(
			fmt.Sprintf("global tasks (max: %d)", m.max),
		)
}

func (m taskListModel) View() string {
	return m.table.View()
}

func (m taskListModel) Pagination() string {
	return m.table.Pagination()
}

func (m taskListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}
