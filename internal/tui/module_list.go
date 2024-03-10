package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

type moduleListModelMaker struct {
	svc        *module.Service
	workspaces *workspace.Service
	workdir    string
}

func (m *moduleListModelMaker) makeModel(_ resource.Resource) (Model, error) {
	columns := []table.Column{
		{Title: "PATH", Width: 30},
		{Title: "CURRENT WORKSPACE", Width: 20},
		{Title: "ID", Width: resource.IDEncodedMaxLen},
	}
	table := table.New[*module.Module](columns).
		WithCellsFunc(func(mod *module.Module) []string {
			return []string{
				mod.Path(),
				mod.Current,
				mod.ID().String(),
			}
		}).
		WithSortFunc(module.ByPath)

	return moduleListModel{
		table:      table,
		svc:        m.svc,
		workspaces: m.workspaces,
		workdir:    m.workdir,
	}, nil
}

type moduleListModel struct {
	table table.Model[*module.Module]

	svc        *module.Service
	workspaces *workspace.Service

	workdir string
}

func (mlm moduleListModel) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[*module.Module](mlm.svc.List())
	}
}

func (m moduleListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Enter):
			if mod, ok := m.table.Highlighted(); ok {
				return m, navigate(page{kind: WorkspaceListKind, resource: mod.Resource})
			}
		case key.Matches(msg, Keys.Init):
			return m, taskCmd(m.svc.Init, maps.Keys(m.table.HighlightedOrSelected())...)
		case key.Matches(msg, Keys.Validate):
			return m, taskCmd(m.svc.Validate, maps.Keys(m.table.HighlightedOrSelected())...)
		case key.Matches(msg, Keys.Format):
			return m, taskCmd(m.svc.Format, maps.Keys(m.table.HighlightedOrSelected())...)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (mlm moduleListModel) Title() string {
	return lipgloss.NewStyle().
		Inherit(Breadcrumbs).
		Padding(0, 0, 0, 1).
		Render(fmt.Sprintf("modules (%s)", mlm.workdir))
}

func (mlm moduleListModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		mlm.table.View(),
	)
}

func (mlm moduleListModel) Footer(width int) string {
	return "logs"
}

func (m moduleListModel) Pagination() string {
	return ""
}

func (mlm moduleListModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}

func (mlm moduleListModel) reload() tea.Cmd {
	return func() tea.Msg {
		if err := mlm.svc.Reload(); err != nil {
			return newErrorMsg(err, "reloading modules")
		}
		return nil
	}
}

func (mlm moduleListModel) reloadWorkspaces(module resource.Resource) tea.Cmd {
	return func() tea.Msg {
		_, err := mlm.workspaces.Reload(module)
		if err != nil {
			return newErrorMsg(err, "reloading workspaces")
		}
		return nil
	}
}
