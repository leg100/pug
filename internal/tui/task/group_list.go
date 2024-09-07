package task

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

var (
	taskGroupCount = table.Column{
		Key:   "tasks",
		Title: "TASKS",
		Width: 10,
	}
	taskGroupID = table.Column{
		Key:   "task_group_id",
		Title: "TASK GROUP ID",
		Width: len("TASK GROUP ID"),
	}
)

type GroupListMaker struct {
	Tasks   *task.Service
	Helpers *tui.Helpers
}

func (m *GroupListMaker) Make(_ resource.ID, width, height int) (tea.Model, error) {
	columns := []table.Column{
		taskGroupID,
		commandColumn,
		taskGroupCount,
		ageColumn,
	}

	renderer := func(g *task.Group) table.RenderedRow {
		row := table.RenderedRow{
			commandColumn.Key:  g.Command,
			taskGroupID.Key:    g.ID.String(),
			taskGroupCount.Key: m.Helpers.GroupReport(g, true),
			ageColumn.Key:      tui.Ago(time.Now(), g.Created),
		}
		return row
	}

	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(task.SortGroupsByCreated),
	)

	return groupList{
		table:   table,
		tasks:   m.Tasks,
		Helpers: m.Helpers,
	}, nil
}

type groupList struct {
	*tui.Helpers

	table table.Model[*task.Group]
	tasks *task.Service
}

func (m groupList) Init() tea.Cmd {
	return func() tea.Msg {
		groups := m.tasks.ListGroups()
		return table.BulkInsertMsg[*task.Group](groups)
	}
}

func (m groupList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, groupListKeys.Enter):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskGroupKind, tui.WithParent(row.ID))
			}
		}
	}
	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m groupList) Title() string {
	return m.Breadcrumbs("TaskGroups", nil)
}

func (m groupList) View() string {
	return m.table.View()
}

func (m groupList) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{}
}
