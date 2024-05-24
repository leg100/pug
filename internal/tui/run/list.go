package run

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/navigator"
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
		table:   table,
		svc:     m.RunService,
		tasks:   m.TaskService,
		parent:  parent,
		helpers: m.Helpers,
	}, nil
}

type list struct {
	table   table.Model[resource.ID, *run.Run]
	svc     tui.RunService
	tasks   tui.TaskService
	parent  resource.Resource
	helpers *tui.Helpers
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
			if row, ok := m.table.CurrentRow(); ok {
				return m, navigator.Go(tui.RunKind, navigator.WithResource(row.Value))
			}
		case key.Matches(msg, keys.Common.Module):
			if row, ok := m.table.CurrentRow(); ok {
				return m, navigator.Go(tui.RunListKind, navigator.WithResource(row.Value.Module()))
			}
		case key.Matches(msg, keys.Common.Workspace):
			if row, ok := m.table.CurrentRow(); ok {
				return m, navigator.Go(tui.RunListKind, navigator.WithResource(row.Value.Workspace()))
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
	return tui.Breadcrumbs("Runs", m.parent, m.table.TotalString())
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
