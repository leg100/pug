package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/workspace"
)

type workspaceListModelMaker struct {
	svc     *workspace.Service
	modules *module.Service
	runs    *run.Service
}

func (m *workspaceListModelMaker) makeModel(parent resource.Resource) (Model, error) {
	var cols []table.Column
	if parent == resource.NilResource {
		// TODO: depending upon kind of parent, hide certain redundant columns, e.g.
		// a module parent kind would render the module column redundant.
		cols = append(cols, table.NewFlexColumn(ColKeyModule, "MODULE", 1).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		))
	}
	cols = append(cols,
		table.NewFlexColumn(ColKeyName, "NAME", 2).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyID, "ID", resource.IDEncodedMaxLen).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	)
	rowMaker := func(ws *workspace.Workspace) table.RowData {
		data := table.RowData{
			ColKeyID:   ws.ID().String(),
			ColKeyName: ws.Workspace().String(),
			ColKeyData: ws,
		}
		// Only show module column if workspaces are not filtered by a parent
		// module.
		if parent == resource.NilResource {
			data[ColKeyModule] = ws.Module().String()
		}
		return data
	}
	return workspaceListModel{
		table: newTableModel(tableModelOptions[*workspace.Workspace]{
			rowMaker: rowMaker,
			columns:  cols,
		}),
		svc:        m.svc,
		modules:    m.modules,
		runs:       m.runs,
		parent:     parent,
		workspaces: make(map[resource.ID]*workspace.Workspace, 0),
	}, nil
}

type workspaceListModel struct {
	table      tableModel[*workspace.Workspace]
	svc        *workspace.Service
	modules    *module.Service
	runs       *run.Service
	parent     resource.Resource
	workspaces map[resource.ID]*workspace.Workspace
}

func (m workspaceListModel) Init() tea.Cmd {
	return func() tea.Msg {
		var opts workspace.ListOptions
		if m.parent != resource.NilResource {
			opts.ModuleID = m.parent.ID()
		}
		return bulkInsertMsg[*workspace.Workspace](m.svc.List(opts))
	}
}

func (m workspaceListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Enter):
			if ok, ws := m.table.highlighted(); ok {
				return m, navigate(page{kind: RunListKind, resource: ws.Resource})
			}
		case key.Matches(msg, Keys.Init):
			return m, taskCmd(m.modules.Init, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, Keys.Format):
			return m, taskCmd(m.modules.Format, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, Keys.Validate):
			return m, taskCmd(m.modules.Validate, m.highlightedOrSelectedModuleIDs()...)
		case key.Matches(msg, Keys.Plan):
			if ok, ws := m.table.highlighted(); ok {
				return m, runCmd(m.runs, ws.ID())
			}
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m workspaceListModel) Title() string {
	var components []string
	if m.parent != resource.NilResource {
		components = append(components,
			lipgloss.NewStyle().
				Background(DarkGrey).
				Foreground(White).
				Bold(true).
				Padding(0, 1).
				Render(m.parent.String()))
	}
	components = append(components,
		lipgloss.NewStyle().
			Background(DarkGrey).
			Foreground(White).
			Bold(true).
			Padding(0, 1).
			Render("workspaces"))
	// Render breadcrumbs together
	return strings.Join(components,
		lipgloss.NewStyle().
			Inherit(Breadcrumbs).
			Render(" › "),
	)
}

func (m workspaceListModel) View() string {
	return m.table.View()
}

func (m workspaceListModel) Pagination() string {
	return m.table.Pagination()
}

func (m workspaceListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}

func (m workspaceListModel) highlightedOrSelectedModuleIDs() []resource.ID {
	selected := m.table.highlightedOrSelected()
	moduleIDs := make([]resource.ID, len(selected))
	for i, s := range selected {
		moduleIDs[i] = s.Module().ID()
	}
	return moduleIDs
}
