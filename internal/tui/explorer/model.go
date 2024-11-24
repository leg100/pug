package explorer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lgtree "github.com/charmbracelet/lipgloss/tree"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

type Maker struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service
	Helpers    *tui.Helpers
	Workdir    internal.Workdir
}

func (m *Maker) Make(_ resource.ID, width, height int) (tea.Model, error) {
	tree := newTree(m.Workdir, nil, nil)
	return model{
		WorkspaceService: m.Workspaces,
		ModuleService:    m.Modules,
		PlanService:      m.Plans,
		Helpers:          m.Helpers,
		Workdir:          m.Workdir,
		tree:             tree,
		tracker:          newTracker(tree),
		w:                width,
		h:                height,
	}, nil
}

type model struct {
	*tui.Helpers

	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	PlanService      *plan.Service
	Workdir          internal.Workdir

	modules    []*module.Module
	workspaces []*workspace.Workspace
	tree       *tree
	tracker    *tracker
	w, h       int
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		return initMsg{
			modules:    m.ModuleService.List(),
			workspaces: m.WorkspaceService.List(workspace.ListOptions{}),
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		createPlanOptions plan.CreateOptions
		applyPrompt       = "Auto-apply %d workspaces?"
		upgrade           bool
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Navigation.LineUp):
			m.tracker.cursorUp()
			return m, nil
		case key.Matches(msg, keys.Navigation.LineDown):
			m.tracker.cursorDown()
			return m, nil
		case key.Matches(msg, keys.Global.Select):
			err := m.tracker.toggleSelection()
			return m, tui.ReportError(err)
		case key.Matches(msg, keys.Global.SelectAll):
			err := m.tracker.selectAll()
			return m, tui.ReportError(err)
		case key.Matches(msg, keys.Global.SelectClear):
			m.tracker.selector.removeAll()
			return m, nil
		case key.Matches(msg, keys.Global.SelectRange):
			err := m.tracker.selectRange()
			return m, tui.ReportError(err)
		case key.Matches(msg, localKeys.Enter):
			m.tracker.toggleClose()
			return m, nil
		case key.Matches(msg, keys.Common.InitUpgrade):
			upgrade = true
			fallthrough
		case key.Matches(msg, keys.Common.Init):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Module {
				return m, tui.ReportError(errors.New("can only trigger init on modules"))
			}
			fn := func(moduleID resource.ID) (task.Spec, error) {
				return m.Modules.Init(moduleID, upgrade)
			}
			return m, m.CreateTasks(fn, ids...)
		case key.Matches(msg, keys.Common.Validate):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Module {
				return m, tui.ReportError(errors.New("can only trigger init on modules"))
			}
			cmd := m.CreateTasks(m.Modules.Validate, ids...)
			return m, cmd
		case key.Matches(msg, keys.Common.Format):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Module {
				return m, tui.ReportError(errors.New("can only trigger format on modules"))
			}
			cmd := m.CreateTasks(m.Modules.Format, ids...)
			return m, cmd
		case key.Matches(msg, localKeys.SetCurrentWorkspace):
			return m, func() tea.Msg {
				currentID := m.tracker.getCursorID()
				if currentID == nil || currentID.Kind != resource.Workspace {
					// TODO: report error
					return nil
				}
				if err := m.Workspaces.SelectWorkspace(*currentID); err != nil {
					return tui.ReportError(fmt.Errorf("setting current workspace: %w", err))()
				}
				return nil
			}
		case key.Matches(msg, keys.Common.PlanDestroy):
			createPlanOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Workspace {
				return m, tui.ReportError(errors.New("can only trigger plans on workspaces"))
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Plan(workspaceID, createPlanOptions)
			}
			return m, m.CreateTasks(fn, ids...)
		case key.Matches(msg, keys.Common.Destroy):
			createPlanOptions.Destroy = true
			applyPrompt = "Destroy resources of %d workspaces?"
			fallthrough
		case key.Matches(msg, keys.Common.Apply):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Workspace {
				return m, tui.ReportError(errors.New("can only trigger applies on workspaces"))
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Apply(workspaceID, createPlanOptions)
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(ids)),
				m.CreateTasks(fn, ids...),
			)
		case key.Matches(msg, keys.Common.Cost):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Workspace {
				return m, tui.ReportError(errors.New("can only trigger infracost on workspaces"))
			}
			spec, err := m.Workspaces.Cost(ids...)
			if err != nil {
				return m, tui.ReportError(fmt.Errorf("creating task: %w", err))
			}
			return m, m.CreateTasksWithSpecs(spec)
		case key.Matches(msg, keys.Common.State, localKeys.Enter):
			currentID := m.tracker.getCursorID()
			if currentID == nil {
				// TODO: report error
				return m, nil
			}
			if currentID.Kind != resource.Workspace {
				// TODO: report error
				return m, nil
			}
			return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(*currentID))
		case key.Matches(msg, keys.Common.Delete):
			workspaceNode, ok := m.tracker.cursorNode.(workspaceNode)
			if !ok {
				// TODO: report error
				return m, nil
			}
			return m, tui.YesNoPrompt(
				fmt.Sprintf("Delete workspace %s?", workspaceNode),
				m.CreateTasks(m.Workspaces.Delete, workspaceNode.id),
			)
		}
	case initMsg:
		m.modules = msg.modules
		m.workspaces = msg.workspaces
		return m, m.buildTree
	case builtTreeMsg:
		m.tree = (*tree)(msg)
		m.tracker.reindex(m.tree)
		return m, nil
	case resource.Event[*module.Module]:
		switch msg.Type {
		case resource.CreatedEvent:
			m.modules = append(m.modules, msg.Payload)
		}
		return m, m.buildTree
	case resource.Event[*workspace.Workspace]:
		switch msg.Type {
		case resource.CreatedEvent:
			m.workspaces = append(m.workspaces, msg.Payload)
		}
		return m, m.buildTree
	case tea.WindowSizeMsg:
		m.w = msg.Width
		m.h = msg.Height
		m.tracker.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	if m.tree == nil {
		return "building tree"
	}
	to := lgtree.New().
		Enumerator(enumerator).
		Indenter(indentor)
	m.tree.render(true, to)
	s := to.String()
	lines := strings.Split(s, "\n")
	totalVisibleLines := min(m.h, len(lines))
	lines = lines[m.tracker.start : m.tracker.start+totalVisibleLines]
	for i := range lines {
		style := lipgloss.NewStyle().Width(m.w - tui.ScrollbarWidth)
		node := m.tracker.nodes[m.tracker.start+i]
		// Style node according to whether it is the cursor node, selected, or
		// both
		if node == m.tracker.cursorNode {
			if m.tracker.isSelected(node) {
				style = style.
					Background(tui.CurrentAndSelectedBackground).
					Foreground(tui.CurrentAndSelectedForeground)
			} else {
				style = style.
					Background(tui.CurrentBackground).
					Foreground(tui.CurrentForeground)
			}
		} else if m.tracker.isSelected(node) {
			style = style.
				Background(tui.SelectedBackground).
				Foreground(tui.SelectedForeground)
		}
		lines[i] = style.Render(lines[i])
	}
	scrollbar := tui.Scrollbar(m.h, len(lines), totalVisibleLines, m.tracker.start)
	return lipgloss.NewStyle().
		Height(m.h).
		MaxHeight(m.h).
		Render(lipgloss.JoinHorizontal(lipgloss.Left,
			strings.Join(lines, "\n"),
			scrollbar,
		))
}

func (m model) Title() string {
	return fmt.Sprintf("start: %d; selected: %v", m.tracker.start, maps.Keys(m.tracker.selections))
}

func (m model) buildTree() tea.Msg {
	tree := newTree(m.Workdir, m.modules, m.workspaces)
	return builtTreeMsg(tree)
}
