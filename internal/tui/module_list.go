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
	"github.com/muesli/reflow/truncateleft"
	"golang.org/x/exp/maps"
)

var moduleColumn = table.Column{
	Title:          "MODULE",
	Width:          20,
	TruncationFunc: truncateleft.StringWithPrefix,
	FlexFactor:     2,
}

type moduleListModelMaker struct {
	svc        *module.Service
	workspaces *workspace.Service
	workdir    string
}

func (m *moduleListModelMaker) makeModel(_ resource.Resource) (Model, error) {
	columns := []table.Column{
		{
			Title:          "MODULE",
			Width:          30,
			TruncationFunc: truncateleft.StringWithPrefix,
			FlexFactor:     1,
		},
		{Title: "CURRENT WORKSPACE", Width: 20},
		{Title: "ID", Width: resource.IDEncodedMaxLen},
	}
	table := table.New[*module.Module](columns).
		WithCellsFunc(func(mod *module.Module) []table.Cell {
			return []table.Cell{
				{Str: mod.Path()},
				{Str: mod.Current},
				{Str: mod.ID().String()},
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
		case key.Matches(msg, Keys.Reload):
			return m, func() tea.Msg {
				if err := m.svc.Reload(); err != nil {
					return newErrorMsg(err, "reloading modules")
				}
				return nil
			}
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

func (m moduleListModel) Pagination() string {
	return ""
}

func (mlm moduleListModel) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		Keys.Init,
		Keys.Validate,
		Keys.Format,
	}
}
