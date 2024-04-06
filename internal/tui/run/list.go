package run

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
)

var ageColumn = table.Column{
	Key:   "age",
	Title: "AGE",
	Width: 10,
}

type ListMaker struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	RunService       tui.RunService
	TaskService      tui.TaskService
	Helpers          *tui.Helpers
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tea.Model, error) {
	var columns []table.Column
	// Add further columns depending upon the kind of parent
	switch parent.Kind {
	case resource.Global:
		// Show module and workspace columns in global runs table
		columns = append(columns, table.ModuleColumn)
		fallthrough
	case resource.Module:
		// Show workspace column in module runs table
		columns = append(columns, table.WorkspaceColumn)
	}
	columns = append(columns,
		table.RunStatusColumn,
		table.RunChangesColumn,
		ageColumn,
		table.IDColumn,
	)

	renderer := func(r *run.Run) table.RenderedRow {
		return table.RenderedRow{
			table.ModuleColumn.Key:     m.Helpers.ModulePath(r.Resource),
			table.WorkspaceColumn.Key:  m.Helpers.WorkspaceName(r.Resource),
			table.RunStatusColumn.Key:  m.Helpers.RunStatus(r),
			table.RunChangesColumn.Key: m.Helpers.LatestRunReport(r),
			ageColumn.Key:              tui.Ago(time.Now(), r.Updated),
			table.IDColumn.Key:         r.String(),
		}
	}
	table := table.NewResource(table.ResourceOptions[*run.Run]{
		Columns:  columns,
		Renderer: renderer,
		Width:    width,
		Height:   height,
		Parent:   parent,
		SortFunc: run.ByStatus,
	})

	return list{
		table:  table,
		svc:    m.RunService,
		tasks:  m.TaskService,
		parent: parent,
	}, nil
}

type list struct {
	table  table.Resource[resource.ID, *run.Run]
	svc    tui.RunService
	tasks  tui.TaskService
	parent resource.Resource
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		runs := m.svc.List(run.ListOptions{
			AncestorID: m.parent.ID,
		})
		return table.BulkInsertMsg[*run.Run](runs)
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
			if row, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.RunKind, tui.WithParent(row.Value.Resource))
			}
		case key.Matches(msg, keys.Common.Cancel):
			// get all highlighted or selected runs, and get the current task
			// for each run, and then cancel those tasks.
		case key.Matches(msg, keys.Common.Apply):
			cmd := tui.CreateTasks("apply", m.svc.Apply, m.table.HighlightedOrSelectedKeys()...)
			m.table.DeselectAll()
			return m, cmd
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.GlobalBreadcrumb("Runs")
}

func (m list) View() string {
	return m.table.View()
}

func (m list) TabStatus() string {
	return fmt.Sprintf("(%d)", len(m.table.Items()))
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		keys.Common.Plan,
		keys.Common.Apply,
		keys.Common.Cancel,
	}
}

//lint:ignore U1000 intend to use shortly
func (m list) navigateLatestTask(runID resource.ID) tea.Cmd {
	return func() tea.Msg {
		tasks := m.tasks.List(task.ListOptions{
			Ancestor: runID,
		})
		var latest *task.Task
		for _, task := range tasks {
			if slices.Equal(task.Command, []string{"apply"}) {
				latest = task
				// Apply task trumps a plan task.
				break
			}
			if slices.Equal(task.Command, []string{"plan"}) {
				latest = task
			}
		}
		if latest == nil {
			return tui.NewErrorMsg(errors.New("no plan or apply task found for run"), "")
		}
		return tui.NewNavigationMsg(tui.TaskKind, tui.WithParent(latest.Resource))
	}
}
