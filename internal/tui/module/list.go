package module

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

// ListMaker makes module list models
type ListMaker struct {
	ModuleService *module.Service
	RunService    *run.Service
	Spinner       *spinner.Model
	Workdir       string
}

func (m *ListMaker) Make(_ resource.Resource, width, height int) (tui.Model, error) {
	currentWorkspace := table.Column{
		Key:        "currentWorkspace",
		Title:      "CURRENT WORKSPACE",
		FlexFactor: 2,
	}
	formatColumn := table.Column{
		Key:   "format",
		Title: "FMAT",
		Width: 4,
	}
	validColumn := table.Column{
		Key:   "valid",
		Title: "VALID",
		Width: 5,
	}
	columns := []table.Column{
		table.ModuleColumn,
		currentWorkspace,
		formatColumn,
		validColumn,
		table.IDColumn,
	}
	boolToUnicode := func(inprog bool, b *bool) string {
		if inprog {
			return m.Spinner.View()
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

	renderer := func(mod *module.Module, style lipgloss.Style) table.RenderedRow {
		cells := table.RenderedRow{
			table.ModuleColumn.Key: mod.Path(),
			formatColumn.Key:       boolToUnicode(mod.FormatInProgress, mod.Formatted),
			validColumn.Key:        boolToUnicode(mod.ValidationInProgress, mod.Valid),
			table.IDColumn.Key:     mod.ID().String(),
		}
		if current := mod.CurrentWorkspace; current != nil {
			cells[currentWorkspace.Key] = current.String()
		}
		return cells
	}
	table := table.New(columns, renderer, width, height)
	table = table.WithSortFunc(module.ByPath)

	return list{
		table:         table,
		spinner:       m.Spinner,
		ModuleService: m.ModuleService,
		RunService:    m.RunService,
		workdir:       m.Workdir,
	}, nil
}

type list struct {
	ModuleService *module.Service
	RunService    *run.Service

	table   table.Model[*module.Module]
	spinner *spinner.Model
	workdir string
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[*module.Module](m.ModuleService.List())
	}
}

func (m list) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case resource.Event[*workspace.Workspace]:
		switch msg.Type {
		case resource.CreatedEvent:
			//cmds = append(cmds, m.createRun(run.CreateOptions{}))
		}
	case resource.Event[*run.Run]:
		//switch msg.Type {
		//case resource.UpdatedEvent:
		//	if msg.Payload.Status == run.Planned {
		//		//				return m, tui.NavigateTo(tui.RunKind, &msg.Payload.Resource)
		//	}
		//}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Enter):
			if mod, ok := m.table.Highlighted(); ok {
				return m, tui.NavigateTo(tui.ModuleKind, &mod.Resource)
			}
		case key.Matches(msg, localKeys.Reload):
			return m, func() tea.Msg {
				if err := m.ModuleService.Reload(); err != nil {
					return tui.NewErrorMsg(err, "reloading modules")
				}
				return nil
			}
		case key.Matches(msg, localKeys.Edit):
			if mod, ok := m.table.Highlighted(); ok {
				// TODO: use env var EDITOR
				// TODO: check for side effects of exec blocking the tui - do
				// messages get queued up?
				c := exec.Command("vim", mod.Path())
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return tui.NewErrorMsg(err, "opening vim")
				})
			}
		case key.Matches(msg, localKeys.Init):
			return m, tui.CreateTasks(m.ModuleService.Init, maps.Keys(m.table.HighlightedOrSelected())...)
		case key.Matches(msg, localKeys.Validate):
			return m, tui.CreateTasks(m.ModuleService.Validate, maps.Keys(m.table.HighlightedOrSelected())...)
		case key.Matches(msg, localKeys.Format):
			return m, tui.CreateTasks(m.ModuleService.Format, maps.Keys(m.table.HighlightedOrSelected())...)
		case key.Matches(msg, localKeys.Plan):
			return m, m.createRun(run.CreateOptions{})
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	workdir := tui.Regular.Copy().Foreground(tui.Blue).Render(m.workdir)
	modules := tui.Bold.Copy().Render("Modules")
	return fmt.Sprintf("%s(%s)", modules, workdir)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) Pagination() string {
	return ""
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return []key.Binding{
		localKeys.Init,
		localKeys.Validate,
		localKeys.Format,
		localKeys.Plan,
		localKeys.Edit,
	}
}

func (m list) createRun(opts run.CreateOptions) tea.Cmd {
	// Handle the case where a user has pressed a key on an empty table with
	// zero rows
	if len(m.table.Items()) == 0 {
		return nil
	}

	// If items have been selected then clear the selection
	var deselectCmd tea.Cmd
	if len(m.table.Selected) > 0 {
		deselectCmd = tui.CmdHandler(table.DeselectMsg{})
	}

	cmd := func() tea.Msg {
		modules := m.table.HighlightedOrSelected()
		var run *run.Run
		for _, mod := range modules {
			if mod.CurrentWorkspace == nil {
				continue
			}
			var err error
			run, err = m.RunService.Create(mod.CurrentWorkspace.ID(), opts)
			if err != nil {
				return tui.NewErrorMsg(err, "creating run")
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
		var target tui.Page
		if len(modules) > 1 {
			target = tui.Page{Kind: tui.RunListKind}
		} else {
			target = tui.Page{Kind: tui.RunKind, Parent: run.Resource}
		}
		return tui.NavigationMsg(target)
	}
	return tea.Batch(cmd, deselectCmd)
}
