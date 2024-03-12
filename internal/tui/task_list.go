package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/table"
	"golang.org/x/exp/maps"
)

type taskListModelMaker struct {
	svc      *task.Service
	maxTasks int
}

func (m *taskListModelMaker) makeModel(parent resource.Resource) (Model, error) {
	// Depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	var (
		columns            []table.Column
		hasModuleColumn    bool
		hasWorkspaceColumn bool
	)
	columns = append(columns, moduleColumn)
	hasModuleColumn = true
	columns = append(columns, table.Column{Title: "WORKSPACE", Width: 10, FlexFactor: 1})
	hasWorkspaceColumn = true
	columns = append(columns,
		table.Column{Title: "COMMAND", Width: 20},
		table.Column{Title: "STATUS", Width: 10},
		table.Column{Title: "AGE", Width: 10},
		table.Column{Title: "ID", Width: resource.IDEncodedMaxLen},
	)
	cellsFunc := func(t *task.Task) []table.Cell {
		cells := make([]table.Cell, 0, len(columns))
		if hasModuleColumn {
			if mod := t.Module(); mod != nil {
				cells = append(cells, table.Cell{Str: mod.String()})
			} else {
				cells = append(cells, table.Cell{})
			}
		}
		if hasWorkspaceColumn {
			if ws := t.Workspace(); ws != nil {
				cells = append(cells, table.Cell{Str: ws.String()})
			} else {
				cells = append(cells, table.Cell{})
			}
		}
		cells = append(cells, table.Cell{Str: strings.Join(t.Command, " ")})

		stateStyle := lipgloss.NewStyle()
		switch t.State {
		case task.Errored:
			stateStyle = stateStyle.Foreground(Red)
		case task.Exited:
			stateStyle = stateStyle.Foreground(lipgloss.Color("40"))
		default:
		}

		cells = append(cells,
			table.Cell{Str: string(t.State), Style: stateStyle},
			table.Cell{Str: ago(time.Now(), t.Updated)},
			table.Cell{Str: t.ID().String()},
		)
		return cells
	}
	table := table.New[*task.Task](columns).
		WithCellsFunc(cellsFunc).
		WithSortFunc(task.ByState)
		// WithWidth(60)

	return taskListModel{
		table:  table,
		svc:    m.svc,
		parent: parent,
		max:    m.maxTasks,
	}, nil
}

type taskListModel struct {
	table  table.Model[*task.Task]
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
		return table.BulkInsertMsg[*task.Task](m.svc.List(opts))
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
			if task, ok := m.table.Highlighted(); ok {
				return m, navigate(page{kind: TaskKind, resource: task.Resource})
			}
		case key.Matches(msg, Keys.Cancel):
			return m, taskCmd(m.svc.Cancel, maps.Keys(m.table.HighlightedOrSelected())...)
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
	return ""
}

func (m taskListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}
