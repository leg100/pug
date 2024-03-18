package tui

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

type moduleListModelMaker struct {
	svc        *module.Service
	workspaces *workspace.Service
	runs       *run.Service

	workdir string

	spinner *spinner.Model
}

func (m *moduleListModelMaker) makeModel(_ resource.Resource) (Model, error) {
	columns := parentColumns(ModuleListKind, resource.Global)
	columns = append(columns,
		table.Column{
			Title:      "CURRENT WORKSPACE",
			FlexFactor: 2,
		},
		table.Column{
			Title: "FMAT",
			Width: 4,
		},
		table.Column{
			Title: "VALID",
			Width: 5,
		},
		table.Column{
			Title: "ID",
			Width: resource.IDEncodedMaxLen,
		},
	)
	boolToUnicode := func(inprog bool, b *bool) string {
		if inprog {
			return m.spinner.View()
		}
		if b == nil {
			return "-"
		}
		if *b {
			return "✓"
		} else {
			return "✗"
		}
	}

	table := table.New[*module.Module](columns).
		WithCellsFunc(func(mod *module.Module) []table.Cell {
			cells := []table.Cell{
				{Str: mod.Path()},
				{},
				{Str: boolToUnicode(mod.FormatInProgress, mod.Formatted)},
				{Str: boolToUnicode(mod.ValidationInProgress, mod.Valid)},
				{Str: mod.ID().String()},
			}
			if current := mod.CurrentWorkspace; current != nil {
				cells[1].Str = current.String()
			}
			return cells
		}).
		WithSortFunc(module.ByPath)

	return moduleListModel{
		table:      table,
		spinner:    m.spinner,
		svc:        m.svc,
		workspaces: m.workspaces,
		runs:       m.runs,
		workdir:    m.workdir,
	}, nil
}

type moduleListModel struct {
	table   table.Model[*module.Module]
	spinner *spinner.Model

	svc        *module.Service
	workspaces *workspace.Service
	runs       *run.Service

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
	case resource.Event[*workspace.Workspace]:
		switch msg.Type {
		case resource.CreatedEvent:
			cmds = append(cmds, m.createRun(run.CreateOptions{}))
		}
	case resource.Event[*run.Run]:
		switch msg.Type {
		case resource.UpdatedEvent:
			if msg.Payload.Status == run.Planned {
				return m, navigate(page{kind: RunKind, resource: msg.Payload.Resource})
			}
		}
	}

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
		case key.Matches(msg, Keys.Edit):
			if mod, ok := m.table.Highlighted(); ok {
				// TODO: use env var EDITOR
				// TODO: check for side effects of exec blocking the tui - do
				// messages get queued up?
				c := exec.Command("vim", mod.Path())
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return newErrorMsg(err, "opening vim")
				})
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
		case key.Matches(msg, Keys.Plan):
			return m, m.createRun(run.CreateOptions{})
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (mlm moduleListModel) Title() string {
	workdir := Regular.Copy().Foreground(Blue).Render(mlm.workdir)
	modules := Bold.Copy().Render("Modules")
	return fmt.Sprintf("%s(%s)", modules, workdir)
}

func (mlm moduleListModel) View() string {
	return mlm.table.View()
}

func (m moduleListModel) Pagination() string {
	return ""
}

func (mlm moduleListModel) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		Keys.Init,
		Keys.Validate,
		Keys.Format,
		Keys.Plan,
		Keys.Edit,
	}
}

func (mlm moduleListModel) createRun(opts run.CreateOptions) tea.Cmd {
	// Handle the case where a user has pressed a key on an empty table with
	// zero rows
	if len(mlm.table.Items()) == 0 {
		return nil
	}

	// If items have been selected then clear the selection
	var deselectCmd tea.Cmd
	if len(mlm.table.Selected) > 0 {
		deselectCmd = cmdHandler(table.DeselectMsg{})
	}

	cmd := func() tea.Msg {
		modules := mlm.table.HighlightedOrSelected()
		var run *run.Run
		for _, mod := range modules {
			if mod.CurrentWorkspace == nil {
				continue
			}
			var err error
			run, err = mlm.runs.Create(mod.CurrentWorkspace.ID(), opts)
			if err != nil {
				return newErrorMsg(err, "creating run")
			}
		}
		if run == nil {
			// No runs were triggered.
			//
			// TODO: show error message in footer
			return nil
		}
		// If user triggered more than one run go to the run listing, otherwise
		// go to the run itself.
		var target page
		if len(modules) > 1 {
			target = page{kind: RunListKind}
		} else {
			target = page{kind: RunKind, resource: run.Resource}
		}
		return navigationMsg{target: target}
	}
	return tea.Batch(cmd, deselectCmd)
}
