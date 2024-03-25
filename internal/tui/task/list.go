package task

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
)

type ListMaker struct {
	TaskService *task.Service
	MaxTasks    int
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tui.Model, error) {
	commandColumn := table.Column{
		Key:        "command",
		Title:      "COMMAND",
		FlexFactor: 1,
	}
	statusColumn := table.Column{
		Key:   "run_status",
		Title: "STATUS",
		Width: run.MaxStatusLen,
	}
	ageColumn := table.Column{
		Key:   "age",
		Title: "AGE",
		Width: 10,
	}
	var columns []table.Column
	switch parent.Kind() {
	case resource.Global:
		// Show all columns in global tasks table
		columns = append(columns, table.ModuleColumn)
		fallthrough
	case resource.Module:
		// Show workspace column in module tasks table
		columns = append(columns, table.WorkspaceColumn)
	}
	columns = append(columns,
		commandColumn,
		statusColumn,
		ageColumn,
	)

	renderer := func(t *task.Task, inherit lipgloss.Style) table.RenderedRow {
		row := table.RenderedRow{
			table.ModuleColumn.Key: t.Module().String(),
			commandColumn.Key:      strings.Join(t.Command, " "),
			ageColumn.Key:          tui.Ago(time.Now(), t.Updated),
			table.IDColumn.Key:     t.ID().String(),
		}
		if ws := t.Workspace(); ws != nil {
			row[table.WorkspaceColumn.Key] = ws.String()
		}

		stateStyle := lipgloss.NewStyle()
		switch t.State {
		case task.Errored:
			stateStyle = stateStyle.Foreground(tui.Red)
		case task.Exited:
			stateStyle = stateStyle.Foreground(lipgloss.Color("40"))
		default:
		}
		row[statusColumn.Key] = stateStyle.Render(string(t.State))
		return row
	}
	table := table.New(columns, renderer, width, height).
		WithSortFunc(task.ByState).
		WithParent(parent)

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
		case key.Matches(msg, keys.Global.Enter):
			if task, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.TaskKind, &task.Resource)
			}
		case key.Matches(msg, keys.Common.Cancel):
			return m, tui.CreateTasks("cancel", m.svc.Cancel, m.table.HighlightedOrSelectedIDs()...)
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

func (m list) TabStatus() string {
	return fmt.Sprintf("(%d)", len(m.table.Items()))
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Cancel,
	}
}
