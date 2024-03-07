package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui/common"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

type workspaceListModelMaker struct {
	svc     *workspace.Service
	modules *module.Service
	runs    *run.Service
}

func (m *workspaceListModelMaker) makeModel(parent resource.Resource) (common.Model, error) {
	// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
	// a module parent kind would render the module column redundant.
	columns := []table.Column{
		table.NewFlexColumn(common.ColKeyModule, "MODULE", 1).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewFlexColumn(common.ColKeyName, "NAME", 2).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}
	return workspaceListModel{
		table:      table.New(columns).Focused(true),
		svc:        m.svc,
		modules:    m.modules,
		runs:       m.runs,
		parent:     parent,
		workspaces: make(map[resource.ID]*workspace.Workspace, 0),
	}, nil
}

type workspaceListModel struct {
	table      table.Model
	svc        *workspace.Service
	modules    *module.Service
	runs       *run.Service
	parent     resource.Resource
	workspaces map[resource.ID]*workspace.Workspace
}

func (m workspaceListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts workspace.ListOptions
		//if m.parent != resource.NilResource {
		//	opts.ModuleID = &m.parent.ID
		//}
		return common.BulkInsertMsg[*workspace.Workspace](m.svc.List(opts))
	}
}

func (m workspaceListModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Enter):
			row := m.table.HighlightedRow()
			ws := row.Data[common.ColKeyData].(*workspace.Workspace)
			return m, navigate(page{kind: RunListKind, resource: ws.Resource})
		case key.Matches(msg, common.Keys.Init):
			row := m.table.HighlightedRow()
			ws := row.Data[common.ColKeyData].(*workspace.Workspace)
			return m, initCmd(m.modules, ws.Module().ID)
		case key.Matches(msg, common.Keys.Plan):
			row := m.table.HighlightedRow()
			ws := row.Data[common.ColKeyData].(*workspace.Workspace)
			return m, runCmd(m.runs, ws.ID)
		case key.Matches(msg, common.Keys.Validate):
			row := m.table.HighlightedRow()
			ws := row.Data[common.ColKeyData].(*workspace.Workspace)
			return m, validateCmd(m.svc, ws.ID)
		case key.Matches(msg, common.Keys.Format):
			row := m.table.HighlightedRow()
			ws := row.Data[common.ColKeyData].(*workspace.Workspace)
			return m, formatCmd(m.svc, ws.ID)
		}
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table = m.table.WithTargetWidth(msg.Width - 2)
		return m, nil
	case common.BulkInsertMsg[*workspace.Workspace]:
		// TODO: filter by parent
		for _, ws := range msg {
			m.workspaces[ws.ID] = ws
		}
		m.table = m.table.WithRows(m.toRows())
	case resource.Event[*workspace.Workspace]:
		// TODO: filter by parent
		switch msg.Type {
		case resource.CreatedEvent:
			m.workspaces[msg.Payload.ID] = msg.Payload
		case resource.UpdatedEvent:
			m.workspaces[msg.Payload.ID] = msg.Payload
		case resource.DeletedEvent:
			delete(m.workspaces, msg.Payload.ID)
		}
		m.table = m.table.WithRows(m.toRows())
		return m, nil
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m workspaceListModel) Title() string {
	// TODO: add optional module
	return "workspaces"
}

func (m workspaceListModel) View() string {
	return m.table.View()
}

func (m workspaceListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

func (m workspaceListModel) toRows() []table.Row {
	rows := make([]table.Row, len(m.workspaces))
	for i, ws := range maps.Values(m.workspaces) {
		rows[i] = table.NewRow(table.RowData{
			common.ColKeyID:     ws.ID.String(),
			common.ColKeyModule: ws.Module().String(),
			common.ColKeyName:   ws.Workspace().String(),
			common.ColKeyData:   ws,
		})
	}
	return rows
}

func runCmd(runs *run.Service, workspaceID resource.ID) tea.Cmd {
	return func() tea.Msg {
		_, task, err := runs.Create(workspaceID, run.CreateOptions{})
		if err != nil {
			return common.NewErrorCmd(err, "creating run")
		}
		return navigationMsg{
			target: page{kind: TaskKind, resource: task.Resource},
		}
	}
}

func validateCmd(workspaces *workspace.Service, workspaceID resource.ID) tea.Cmd {
	return func() tea.Msg {
		task, err := workspaces.Validate(workspaceID)
		if err != nil {
			return common.NewErrorCmd(err, "creating validate task")
		}
		return navigationMsg{
			target: page{kind: TaskKind, resource: task.Resource},
		}
	}
}

func formatCmd(workspaces *workspace.Service, workspaceID resource.ID) tea.Cmd {
	return func() tea.Msg {
		task, err := workspaces.Format(workspaceID)
		if err != nil {
			return common.NewErrorCmd(err, "creating format task")
		}
		return navigationMsg{
			target: page{kind: TaskKind, resource: task.Resource},
		}
	}
}
