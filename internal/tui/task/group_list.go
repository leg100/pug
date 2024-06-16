package task

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
)

var (
	taskGroupCount = table.Column{
		Key:   "tasks",
		Title: "TASKS",
		Width: len("TASKS"),
	}
	taskGroupID = table.Column{
		Key:   "task_group_id",
		Title: "TASK GROUP ID",
		Width: resource.IDEncodedMaxLen,
	}
)

type GroupListMaker struct {
	TaskService tui.TaskService
	Helpers     *tui.Helpers
}

func (m *GroupListMaker) Make(_ resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{
		commandColumn,
		taskGroupID,
		taskGroupCount,
		ageColumn,
	}

	renderer := func(g *task.Group) table.RenderedRow {
		row := table.RenderedRow{
			commandColumn.Key:  g.Command,
			taskGroupID.Key:    g.ID.String(),
			taskGroupCount.Key: fmt.Sprintf("%d/%d", g.Finished(), len(g.Tasks)),
			ageColumn.Key:      tui.Ago(time.Now(), g.Created),
		}
		return row
	}

	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(task.SortGroupsByCreated),
	)

	return groupList{
		table: table,
		svc:   m.TaskService,
	}, nil
}

type groupList struct {
	table table.Model[*task.Group]
	svc   tui.TaskService
}

func (m groupList) Init() tea.Cmd {
	return func() tea.Msg {
		groups := m.svc.ListGroups()
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
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskGroupKind, tui.WithParent(row.Value))
			}
		}
	}
	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m groupList) Title() string {
	return tui.Breadcrumbs("TaskGroups", resource.GlobalResource)
}

func (m groupList) View() string {
	return m.table.View()
}

func (m groupList) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{}
}
