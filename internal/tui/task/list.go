package task

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
)

var (
	commandColumn = table.Column{
		Key:        "command",
		Title:      "COMMAND",
		FlexFactor: 1,
	}
	statusColumn = table.Column{
		Key:   "run_status",
		Title: "STATUS",
		Width: run.MaxStatusLen,
	}
	ageColumn = table.Column{
		Key:   "age",
		Title: "AGE",
		Width: 10,
	}
)

type ListMaker struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
	TaskService      *task.Service
	MaxTasks         int
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tui.Model, error) {
	var columns []table.Column
	// Add further columns depending upon the kind of parent
	switch parent.Kind {
	case resource.Global:
		// Show module and workspace columns in global tasks table
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
			commandColumn.Key:  t.CommandString(),
			ageColumn.Key:      tui.Ago(time.Now(), t.Updated),
			table.IDColumn.Key: t.String(),
		}
		switch parent.Kind {
		case resource.Global:
			// Show module path and workspace name in global tasks table if it
			// has one.
			if mod := t.Module(); mod != nil {
				mod, err := m.ModuleService.Get(mod.ID)
				if err != nil {
					slog.Error("rendering module path", "error", err)
					break
				}
				row[table.ModuleColumn.Key] = mod.Path
			}
			fallthrough
		case resource.Module:
			// Show workspace name in module tasks table if it has one.
			if ws := t.Workspace(); ws != nil {
				ws, err := m.WorkspaceService.Get(ws.ID)
				if err != nil {
					slog.Error("rendering workspace name", "error", err)
					break
				}
				row[table.WorkspaceColumn.Key] = ws.Name
			}
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

	table := table.NewResource(table.ResourceOptions[*task.Task]{
		ModuleService:    m.ModuleService,
		WorkspaceService: m.WorkspaceService,
		Columns:          columns,
		Renderer:         renderer,
		Width:            width,
		Height:           height,
		Parent:           parent,
		SortFunc:         task.ByState,
	})

	return list{
		table:  table,
		svc:    m.TaskService,
		parent: parent,
		max:    m.MaxTasks,
	}, nil
}

type list struct {
	table  table.Resource[resource.ID, *task.Task]
	svc    *task.Service
	parent resource.Resource
	max    int
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := m.svc.List(task.ListOptions{
			Ancestor: m.parent.ID,
		})
		return table.BulkInsertMsg[*task.Task](tasks)
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
			if row, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.TaskKind, tui.WithParent(row.Value.Resource))
			}
		case key.Matches(msg, keys.Common.Cancel):
			return m, tui.CreateTasks("cancel", m.svc.Cancel, m.table.HighlightedOrSelectedKeys()...)
		}
	}
	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.GlobalBreadcrumb("Tasks")
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
