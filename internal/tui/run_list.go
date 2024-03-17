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
	"github.com/leg100/pug/internal/tui/table"
)

type runListModelMaker struct {
	svc   *run.Service
	tasks *task.Service
}

func (m *runListModelMaker) makeModel(parent resource.Resource) (Model, error) {
	columns := append(parentColumns(RunListKind, parent.Kind),
		table.Column{Title: "STATUS", Width: run.MaxStatusLen},
		table.Column{Title: "CHANGES", Width: 14},
		table.Column{Title: "AGE", Width: 10},
	)
	cellsFunc := func(r *run.Run) []table.Cell {
		cells := parentCells(RunListKind, parent.Kind, r.Resource)
		cells = append(cells, table.Cell{Str: string(r.Status)})

		// switch r.Status {
		// case run.Planned, run.PlannedAndFinished, run.ApplyQueued, run.Applying:
		// case run.Applied:
		// }
		cells = append(cells, table.Cell{Str: r.PlanReport.String()})
		cells = append(cells, table.Cell{Str: ago(time.Now(), r.Created)})

		return cells
	}
	table := table.New[*run.Run](columns).
		WithCellsFunc(cellsFunc).
		WithSortFunc(run.ByUpdatedDesc).
		WithParent(parent)

	return runListModel{
		table:  table,
		svc:    m.svc,
		tasks:  m.tasks,
		parent: parent,
	}, nil
}

type runListModel struct {
	table  table.Model[*run.Run]
	svc    *run.Service
	tasks  *task.Service
	parent resource.Resource
}

func (m runListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts run.ListOptions
		if m.parent != resource.GlobalResource {
			opts.ParentID = m.parent.ID()
		}
		return table.BulkInsertMsg[*run.Run](m.svc.List(opts))
	}
}

func (m runListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Enter):
			if run, ok := m.table.Highlighted(); ok {
				return m, navigate(page{kind: TaskListKind, resource: run.Resource})
			}
		case key.Matches(msg, Keys.Cancel):
			// get all highlighted or selected runs, and get the current task
			// for each run, and then cancel those tasks.
		case key.Matches(msg, Keys.Apply):
			hl, ok := m.table.Highlighted()
			if !ok {
				return m, nil
			}
			if hl.Status != run.Planned {
				return m, nil
			}
			cmds = append(cmds, func() tea.Msg {
				_, task, err := m.svc.Apply(hl.ID())
				if err != nil {
					return newErrorMsg(err, "creating apply task")
				}
				return navigationMsg{
					target: page{kind: TaskKind, resource: task.Resource},
				}
			})
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m runListModel) Title() string {
	return breadcrumbs("Runs", m.parent)
}

func (m runListModel) View() string {
	return m.table.View()
}

func (m runListModel) Pagination() string {
	return ""
}

func (m runListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings,
		Keys.Plan,
		Keys.Apply,
		Keys.Cancel,
	)
	return
}

//lint:ignore U1000 intend to use shortly
func (m runListModel) navigateLatestTask(runID resource.ID) tea.Cmd {
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
			return newErrorMsg(errors.New("no plan or apply task found for run"), "")
		}
		return navigationMsg{page{kind: TaskKind, resource: latest.Resource}}
	}
}

func runCmd(runs *run.Service, workspaceID resource.ID) tea.Cmd {
	return func() tea.Msg {
		_, _, err := runs.Create(workspaceID, run.CreateOptions{})
		if err != nil {
			return newErrorMsg(err, "creating run")
		}
		return navigationMsg{
			target: page{kind: TaskListKind},
		}
	}
}
