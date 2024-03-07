package tui

import (
	"errors"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
	"golang.org/x/exp/maps"
)

type runListModelMaker struct {
	svc   *run.Service
	tasks *task.Service
}

func (m *runListModelMaker) makeModel(parent resource.Resource) (common.Model, error) {
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		table.NewColumn(common.ColKeyID, "ID", 10),
		table.NewColumn(common.ColKeyModule, "MODULE", 10),
		table.NewColumn(common.ColKeyWorkspace, "WORKSPACE", 10),
		table.NewColumn(common.ColKeyStatus, "STATUS", 10),
		table.NewColumn(common.ColKeyAgo, "AGE", 10),
	}
	return runListModel{
		table:  table.New(columns),
		svc:    m.svc,
		tasks:  m.tasks,
		parent: parent,
		runs:   make(map[resource.ID]*run.Run, 0),
	}, nil
}

type runListModel struct {
	table  table.Model
	svc    *run.Service
	tasks  *task.Service
	parent resource.Resource

	runs map[resource.ID]*run.Run
}

func (m runListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts run.ListOptions
		if m.parent != resource.NilResource {
			opts.ParentID = &m.parent.ID
		}
		return common.BulkInsertMsg[*run.Run](m.svc.List(opts))
	}
}

func (m runListModel) toRows() []table.Row {
	rows := make([]table.Row, len(m.runs))
	for i, run := range maps.Values(m.runs) {
		rows[i] = table.NewRow(table.RowData{
			common.ColKeyID:        run.ID.String(),
			common.ColKeyModule:    run.Module().String(),
			common.ColKeyWorkspace: run.Workspace().String(),
			common.ColKeyStatus:    string(run.Status),
			common.ColKeyAgo:       run.Created.Round(time.Second).String(),
			common.ColKeyData:      run,
		})
	}
	return rows
}

func (m runListModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Enter):
			// When user presses enter on a run, then it should navigate to the
			// task for its plan if only a plan has been run, or to the task for
			// its apply, if that has been run.
			row := m.table.HighlightedRow()
			run := row.Data[common.ColKeyData].(*run.Run)
			return m, m.navigateLatestTask(run.ID)
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table = m.table.WithMaxTotalWidth(msg.Width - 2)
		m.table = m.table.WithMinimumHeight(msg.Height)
		return m, nil
	case common.BulkInsertMsg[*run.Run]:
		m.runs = make(map[resource.ID]*run.Run, len(msg))
		for _, run := range msg {
			m.runs[run.ID] = run
		}
		m.table = m.table.WithRows(m.toRows())
	case resource.Event[*run.Run]:
		switch msg.Type {
		case resource.CreatedEvent:
			m.runs[msg.Payload.ID] = msg.Payload
		case resource.UpdatedEvent:
			m.runs[msg.Payload.ID] = msg.Payload
		case resource.DeletedEvent:
			delete(m.runs, msg.Payload.ID)
		}
		m.table = m.table.WithRows(m.toRows())
		return m, nil
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m runListModel) Title() string {
	return "runs"
}

func (m runListModel) View() string {
	return m.table.View()
}

func (m runListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
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
			return common.NewErrorCmd(errors.New("no plan or apply task found for run"), "")
		}
		return navigationMsg{page{kind: TaskKind, resource: latest.Resource}}
	}
}
