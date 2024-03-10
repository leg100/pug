package tui

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
)

type runListModelMaker struct {
	svc   *run.Service
	tasks *task.Service
}

func (m *runListModelMaker) makeModel(parent resource.Resource) (Model, error) {
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	cols := []table.Column{
		table.NewFlexColumn(ColKeyModule, "MODULE", 3),
		table.NewFlexColumn(ColKeyWorkspace, "WORKSPACE", 2),
		table.NewColumn(ColKeyStatus, "STATUS", 20),
		table.NewColumn(ColKeyAgo, "AGE", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyID, "ID", resource.IDEncodedMaxLen),
	}
	rowMaker := func(run *run.Run) table.RowData {
		data := table.RowData{
			ColKeyID:     run.ID().String(),
			ColKeyModule: run.Module().String(),
			ColKeyStatus: string(run.Status),
			ColKeyAgo:    ago(time.Now(), run.Created),
		}
		// Only show module column if not filtered by a parent module.
		if parent == resource.NilResource {
			data[ColKeyModule] = run.Module().String()
		}
		// Only show workspace column if not filtered by a parent workspace.
		if parent == resource.NilResource {
			data[ColKeyWorkspace] = run.Workspace().String()
		}
		return data
	}
	return runListModel{
		table: newTableModel(tableModelOptions[*run.Run]{
			rowMaker: rowMaker,
			columns:  cols,
		}),
		svc:    m.svc,
		tasks:  m.tasks,
		parent: parent,
	}, nil
}

type runListModel struct {
	table  tableModel[*run.Run]
	svc    *run.Service
	tasks  *task.Service
	parent resource.Resource
}

func (m runListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts run.ListOptions
		if m.parent != resource.NilResource {
			opts.ParentID = m.parent.ID()
		}
		return bulkInsertMsg[*run.Run](m.svc.List(opts))
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
			// When user presses enter on a run, then it should navigate to the
			// task for its plan if only a plan has been run, or to the task for
			// its apply, if that has been run.
			if ok, run := m.table.highlighted(); ok {
				return m, m.navigateLatestTask(run.ID())
			}
		case key.Matches(msg, Keys.Cancel):
			// get all highlighted or selected runs, and get the current task
			// for each run, and then cancel those tasks.
		case key.Matches(msg, Keys.Apply):
			ok, hl := m.table.highlighted()
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
	return lipgloss.NewStyle().
		Inherit(Breadcrumbs).
		Padding(0, 0, 0, 1).
		Render(
			fmt.Sprintf("runs"),
		)
}

func (m runListModel) View() string {
	return m.table.View()
}

func (m runListModel) Pagination() string {
	return m.table.Pagination()
}

func (m runListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings,
		Keys.Plan,
		Keys.Apply,
		Keys.Cancel,
	)
	return
}

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
		return nil
		//return navigationMsg{
		//	target: page{kind: TaskKind, resource: task.Resource},
		//}
	}
}
