package tui

import (
	"errors"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

type ListMaker struct {
	RunService  *run.Service
	TaskService *task.Service
}

func (m *ListMaker) Make(parent resource.Resource, width, height int) (tui.Model, error) {
	columns := append(tui.ParentColumns(tui.RunListKind, parent.Kind),
		table.Column{Title: "STATUS", Width: run.MaxStatusLen},
		table.Column{Title: "CHANGES", Width: 14},
		table.Column{Title: "AGE", Width: 10},
	)
	cellsFunc := func(r *run.Run) []table.Cell {
		cells := tui.ParentCells(tui.RunListKind, parent.Kind, r.Resource)
		cells = append(cells, table.Cell{Str: string(r.Status)})

		// switch r.Status {
		// case run.Planned, run.PlannedAndFinished, run.ApplyQueued, run.Applying:
		// case run.Applied:
		// }
		cells = append(cells, table.Cell{Str: r.PlanReport.String()})
		cells = append(cells, table.Cell{Str: tui.Ago(time.Now(), r.Created)})

		return cells
	}
	table := table.New[*run.Run](columns).
		WithCellsFunc(cellsFunc).
		WithSortFunc(run.ByUpdatedDesc).
		WithParent(parent)

	return list{
		table:  table,
		svc:    m.RunService,
		tasks:  m.TaskService,
		parent: parent,
	}, nil
}

type list struct {
	table  table.Model[*run.Run]
	svc    *run.Service
	tasks  *task.Service
	parent resource.Resource
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		var opts run.ListOptions
		if m.parent != resource.GlobalResource {
			opts.AncestorID = m.parent.ID()
		}
		return table.BulkInsertMsg[*run.Run](m.svc.List(opts))
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
			if run, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.RunKind, &run.Resource)
			}
		case key.Matches(msg, tui.Keys.Cancel):
			// get all highlighted or selected runs, and get the current task
			// for each run, and then cancel those tasks.
		case key.Matches(msg, tui.Keys.Apply):
			hl, ok := m.table.Highlighted()
			if !ok {
				return m, nil
			}
			if hl.Status != run.Planned {
				return m, nil
			}
			cmds = append(cmds, func() tea.Msg {
				task, err := m.svc.Apply(hl.ID())
				if err != nil {
					return tui.NewErrorMsg(err, "creating apply task")
				}
				return tui.NavigationMsg(
					tui.Page{Kind: tui.TaskKind, Parent: task.Resource},
				)
			})
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.Breadcrumbs("Runs", m.parent)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) Pagination() string {
	return ""
}

func (m list) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings,
		tui.Keys.Plan,
		tui.Keys.Apply,
		tui.Keys.Cancel,
	)
	return
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
		return tui.NavigationMsg(tui.Page{Kind: tui.TaskKind, Parent: latest.Resource})
	}
}
