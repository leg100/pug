package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
	"golang.org/x/exp/maps"
)

type ListMaker struct {
	TaskService *task.Service
	MaxTasks    int
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tui.Model, error) {
	columns := append(tui.ParentColumns(tui.TaskListKind, parent.Kind),
		table.Column{Title: "COMMAND", Width: 20},
		table.Column{Title: "STATUS", Width: 10},
		table.Column{Title: "AGE", Width: 10},
	)
	cellsFunc := func(t *task.Task) []table.Cell {
		cells := tui.ParentCells(tui.TaskListKind, parent.Kind, t.Resource)
		cells = append(cells, table.Cell{Str: strings.Join(t.Command, " ")})

		stateStyle := lipgloss.NewStyle()
		switch t.State {
		case task.Errored:
			stateStyle = stateStyle.Foreground(tui.Red)
		case task.Exited:
			stateStyle = stateStyle.Foreground(lipgloss.Color("40"))
		default:
		}

		cells = append(cells,
			table.Cell{Str: string(t.State), Style: stateStyle},
			table.Cell{Str: tui.Ago(time.Now(), t.Updated)},
		)
		return cells
	}
	table := table.New[*task.Task](columns).
		WithCellsFunc(cellsFunc).
		WithSortFunc(task.ByState).
		WithParent(parent)
		// WithWidth(60)

	return list{
		table:  table,
		svc:    m.TaskService,
		parent: parent,
		max:    m.MaxTasks,
	}, nil
}

type list struct {
	table  table.Model[*task.Task]
	svc    *task.Service
	parent resource.Resource
	max    int
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		var opts task.ListOptions
		if m.parent != resource.GlobalResource {
			opts.Ancestor = m.parent.ID()
		}
		return table.BulkInsertMsg[*task.Task](m.svc.List(opts))
	}
}

func (m list) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Enter):
			if task, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.TaskKind, &task.Resource)
			}
		case key.Matches(msg, tui.Keys.Cancel):
			return m, TaskCmd(m.svc.Cancel, maps.Keys(m.table.HighlightedOrSelected())...)
		}
	}
	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.Breadcrumbs(tui.Bold.Render("Tasks"), m.parent)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) Pagination() string {
	return ""
}

func (m list) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, tui.Keys.CloseHelp)
	return
}
