package taskgroup

import (
	"strconv"
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
	commandColumn = table.Column{
		Key:        "command",
		Title:      "COMMAND",
		FlexFactor: 1,
	}
	ageColumn = table.Column{
		Key:   "age",
		Title: "AGE",
		Width: 7,
	}
	taskGroupCount = table.Column{
		Key:   "tasks",
		Title: "TASKS",
		Width: len("TASKS"),
	}
)

type ListMaker struct {
	TaskService tui.TaskService
	Helpers     *tui.Helpers
}

func (m *ListMaker) Make(_ resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{
		commandColumn,
		taskGroupCount,
		ageColumn,
	}

	renderer := func(g *task.Group) table.RenderedRow {
		row := table.RenderedRow{
			commandColumn.Key:  g.Command,
			taskGroupCount.Key: strconv.Itoa(len(g.Tasks)),
			ageColumn.Key:      tui.Ago(time.Now(), g.Created),
		}
		return row
	}

	table := table.New(columns, renderer, width, height).
		WithSortFunc(task.SortGroupsByCreated)

	return list{
		table: table,
		svc:   m.TaskService,
	}, nil
}

type list struct {
	table table.Model[resource.ID, *task.Group]
	svc   tui.TaskService
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		groups := m.svc.ListGroups()
		return table.BulkInsertMsg[*task.Group](groups)
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m list) Title() string {
	return tui.GlobalBreadcrumb("TaskGroups", m.table.TotalString())
}

func (m list) View() string {
	return m.table.View()
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{}
}
