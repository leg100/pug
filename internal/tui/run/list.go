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
	switch parent.GetKind() {
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
			table.ModuleColumn.Key:     r.ModulePath(),
			table.WorkspaceColumn.Key:  r.WorkspaceName(),
			table.RunStatusColumn.Key:  m.Helpers.RunStatus(r),
			table.RunChangesColumn.Key: m.Helpers.LatestRunReport(r),
			ageColumn.Key:              tui.Ago(time.Now(), r.Updated),
			table.IDColumn.Key:         r.String(),
		}
	}
	table := table.New(columns, renderer, width, height).
		WithSortFunc(run.ByStatus).
		WithParent(parent)

	return list{
		table:  table,
		svc:    m.RunService,
		tasks:  m.TaskService,
		parent: parent,
	}, nil
}

type list struct {
	table  table.Model[resource.ID, *run.Run]
	svc    tui.RunService
	tasks  tui.TaskService
	parent resource.Resource
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		runs := m.svc.List(run.ListOptions{
			AncestorID: m.parent.GetID(),
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
			if row, highlighted := m.table.Highlighted(); highlighted {
				return m, tui.NavigateTo(tui.RunKind, tui.WithParent(row.Value))
			}
		case key.Matches(msg, keys.Common.Apply):
			runIDs, err := m.table.Prune(func(row *run.Run) (resource.ID, error) {
				if row.Status != run.Planned {
					return resource.ID{}, errors.New("run is not in the planned state")
				}
				return row.ID, nil
			})
			if err != nil {
				return m, tui.ReportError(err, "")
			}
			return m, ApplyCommand(m.svc, m.parent, runIDs...)
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.GlobalBreadcrumb("Runs", m.table.TotalString())
}

func (m list) View() string {
	return m.table.View()
}

func (m list) TabStatus() string {
	return fmt.Sprintf("(%s)", m.table.TotalString())
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
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
		return tui.NewNavigationMsg(tui.TaskKind, tui.WithParent(latest))
	}
}
