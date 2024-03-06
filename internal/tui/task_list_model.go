package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/evertras/bubble-table/table"
	"golang.org/x/exp/maps"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
)

type taskListModelMaker struct {
	svc      *task.Service
	maxTasks int
}

func (m *taskListModelMaker) makeModel(parent resource.Resource) (common.Model, error) {
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		table.NewColumn(common.ColKeyID, "ID", resource.IDEncodedMaxLen).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewFlexColumn(common.ColKeyModule, "MODULE", 2).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(common.ColKeyWorkspace, "WORKSPACE", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewFlexColumn(common.ColKeyCommand, "COMMAND", 1).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(common.ColKeyStatus, "STATUS", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(common.ColKeyAge, "AGE", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}

	return taskListModel{
		table:  table.New(columns).Focused(true).SortByDesc(common.ColKeyTime),
		svc:    m.svc,
		parent: parent,
		tasks:  make(map[resource.ID]*task.Task),
		max:    m.maxTasks,
	}, nil
}

type taskListModel struct {
	table  table.Model
	svc    *task.Service
	tasks  map[resource.ID]*task.Task
	parent resource.Resource
	max    int
}

func (m taskListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts task.ListOptions
		if m.parent != resource.NilResource {
			opts.Ancestor = m.parent.ID
		}
		return common.BulkInsertMsg[*task.Task](m.svc.List(opts))
	}
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
			row := m.table.HighlightedRow()
			task := row.Data[common.ColKeyData].(*task.Task)
			return m, navigate(page{kind: TaskKind, resource: task.Resource})
			//case key.Matches(msg, common.Keys.Cancel):
			//	row := m.table.HighlightedRow()
			//	task := row.Data[common.ColKeyData].(*task.Task)
			//	return m, cancelCmd(m.svc, task.ID)
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table = m.table.WithTargetWidth(msg.Width - 2)
		m.table = m.table.WithMinimumHeight(msg.Height)
	case common.BulkInsertMsg[*task.Task]:
		m.tasks = make(map[resource.ID]*task.Task, len(msg))
		for _, run := range msg {
			m.tasks[run.ID] = run
		}
		m.table = m.table.WithRows(m.toRows())
	case resource.Event[*task.Task]:
		switch msg.Type {
		case resource.CreatedEvent:
			m.tasks[msg.Payload.ID] = msg.Payload
		case resource.UpdatedEvent:
			m.tasks[msg.Payload.ID] = msg.Payload
		case resource.DeletedEvent:
			delete(m.tasks, msg.Payload.ID)
		}
		m.table = m.table.WithRows(m.toRows())
		return m, nil
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m taskListModel) Title() string {
	return fmt.Sprintf("global tasks (max: %d)", m.max)
}

func (m taskListModel) View() string {
	return m.table.View()
}

func (m taskListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

func (m taskListModel) toRows() []table.Row {
	rows := make([]table.Row, len(m.tasks))
	// TODO: categorise and sort tasks according to a specific schema:
	// 1. running (newest first)
	// 2. queued (oldest first)
	// 3. pending (oldest first)
	// 4. finished (newest first)
	for i, task := range maps.Values(m.tasks) {
		row := table.RowData{
			common.ColKeyID:      task.ID.String(),
			common.ColKeyCommand: strings.Join(task.Command, " "),
			common.ColKeyStatus:  string(task.State),
			common.ColKeyAge:     ago(time.Now(), task.Created),
			common.ColKeyData:    task,
		}
		if mod := task.Module(); mod != nil {
			row[common.ColKeyModule] = mod.String()
		}
		if ws := task.Workspace(); ws != nil {
			row[common.ColKeyWorkspace] = ws.String()
		}
		rows[i] = table.NewRow(row)
	}
	return rows
}

func cancelCmd(tasks *task.Service, taskID resource.ID) tea.Cmd {
	return func() tea.Msg {
		task, err := tasks.Cancel(taskID)
		if err != nil {
			return common.NewErrorCmd(err, "creating cancel task")
		}
		return navigationMsg{
			target: page{kind: TaskKind, resource: task.Resource},
		}
	}
}
