package explorer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lgtree "github.com/charmbracelet/lipgloss/tree"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/workspace"
)

type Maker struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	PlanService      *plan.Service
	Workdir          internal.Workdir
	Helpers          *tui.Helpers
}

func (m *Maker) Make(id resource.ID, width, height int) (tui.ChildModel, error) {
	builder := &treeBuilder{
		wd:               m.Workdir,
		helpers:          m.Helpers,
		moduleService:    m.ModuleService,
		workspaceService: m.WorkspaceService,
	}
	tree := builder.newTree("")
	filter := textinput.New()
	filter.Prompt = "Filter: "
	return &model{
		WorkspaceService: m.WorkspaceService,
		ModuleService:    m.ModuleService,
		PlanService:      m.PlanService,
		Helpers:          m.Helpers,
		Workdir:          m.Workdir,
		treeBuilder:      builder,
		tree:             tree,
		tracker:          newTracker(tree, height),
		filter:           filter,
	}, nil
}

type model struct {
	*tui.Helpers

	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	PlanService      *plan.Service
	Workdir          internal.Workdir

	treeBuilder   *treeBuilder
	tree          *tree
	tracker       *tracker
	width, height int
	filter        textinput.Model
}

func (m model) Init() tea.Cmd {
	return m.buildTree
}

func (m *model) Update(msg tea.Msg) tea.Cmd {
	var (
		createPlanOptions plan.CreateOptions
		applyPrompt       = "Auto-apply %d workspaces?"
		upgrade           bool
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Navigation.LineUp):
			m.tracker.moveCursor(-1, m.treeHeight())
		case key.Matches(msg, keys.Navigation.LineDown):
			m.tracker.moveCursor(1, m.treeHeight())
		case key.Matches(msg, keys.Navigation.PageUp):
			m.tracker.moveCursor(-m.treeHeight(), m.treeHeight())
		case key.Matches(msg, keys.Navigation.PageDown):
			m.tracker.moveCursor(m.treeHeight(), m.treeHeight())
		case key.Matches(msg, keys.Navigation.HalfPageUp):
			m.tracker.moveCursor(-m.treeHeight()/2, m.treeHeight())
		case key.Matches(msg, keys.Navigation.HalfPageDown):
			m.tracker.moveCursor(m.treeHeight()/2, m.treeHeight())
		case key.Matches(msg, keys.Navigation.GotoTop):
			m.tracker.moveCursor(-m.tracker.cursorIndex, m.treeHeight())
		case key.Matches(msg, keys.Navigation.GotoBottom):
			m.tracker.moveCursor(len(m.tracker.nodes), m.treeHeight())
		case key.Matches(msg, keys.Global.Select):
			err := m.tracker.toggleSelection()
			return tui.ReportError(err)
		case key.Matches(msg, keys.Global.SelectAll):
			err := m.tracker.selectAll()
			return tui.ReportError(err)
		case key.Matches(msg, keys.Global.SelectClear):
			m.tracker.selector.removeAll()
			return nil
		case key.Matches(msg, keys.Global.SelectRange):
			err := m.tracker.selectRange()
			return tui.ReportError(err)
		case key.Matches(msg, localKeys.Enter):
			m.tracker.toggleClose()
			return nil
		case key.Matches(msg, keys.Common.InitUpgrade):
			upgrade = true
			fallthrough
		case key.Matches(msg, keys.Common.Init):
			ids, err := m.getModuleIDs()
			if err != nil {
				return tui.ReportError(err)
			}
			fn := func(moduleID resource.ID) (task.Spec, error) {
				return m.Modules.Init(moduleID, upgrade)
			}
			return m.CreateTasks(fn, ids...)
		case key.Matches(msg, keys.Common.Validate):
			ids, err := m.getModuleIDs()
			if err != nil {
				return tui.ReportError(err)
			}
			cmd := m.CreateTasks(m.Modules.Validate, ids...)
			return cmd
		case key.Matches(msg, keys.Common.Format):
			ids, err := m.getModuleIDs()
			if err != nil {
				return tui.ReportError(err)
			}
			cmd := m.CreateTasks(m.Modules.Format, ids...)
			return cmd
		case key.Matches(msg, localKeys.SetCurrentWorkspace):
			return func() tea.Msg {
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
			ids, err := m.getWorkspaceIDs()
			if err != nil {
				return tui.ReportError(err)
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Plan(workspaceID, createPlanOptions)
			}
			return m.CreateTasks(fn, ids...)
		case key.Matches(msg, keys.Common.Destroy):
			createPlanOptions.Destroy = true
			applyPrompt = "Destroy resources of %d workspaces?"
			fallthrough
		case key.Matches(msg, keys.Common.Apply):
			ids, err := m.getWorkspaceIDs()
			if err != nil {
				return tui.ReportError(err)
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Apply(workspaceID, createPlanOptions)
			}
			return tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(ids)),
				m.CreateTasks(fn, ids...),
			)
		case key.Matches(msg, keys.Common.Cost):
			ids, err := m.getWorkspaceIDs()
			if err != nil {
				return tui.ReportError(err)
			}
			spec, err := m.Workspaces.Cost(ids...)
			if err != nil {
				return tui.ReportError(fmt.Errorf("creating task: %w", err))
			}
			return m.CreateTasksWithSpecs(spec)
		case key.Matches(msg, keys.Common.State, localKeys.Enter):
			currentID := m.tracker.getCursorID()
			if currentID == nil {
				// TODO: report error
				return nil
			}
			if currentID.Kind != resource.Workspace {
				// TODO: report error
				return nil
			}
			return tui.NavigateTo(tui.ResourceListKind, tui.WithParent(*currentID))
		case key.Matches(msg, keys.Common.Delete):
			workspaceNode, ok := m.tracker.cursorNode.(workspaceNode)
			if !ok {
				// TODO: report error
				return nil
			}
			return tui.YesNoPrompt(
				fmt.Sprintf("Delete workspace %s?", workspaceNode),
				m.CreateTasks(m.Workspaces.Delete, workspaceNode.id),
			)
		}
	case builtTreeMsg:
		m.tree = (*tree)(msg)
		// TODO: perform this in a cmd
		m.tracker.reindex(m.tree, m.treeHeight())
		return nil
	case resource.Event[*module.Module]:
		return m.buildTree
	case resource.Event[*workspace.Workspace]:
		return m.buildTree
	case resource.Event[*state.State]:
		return m.buildTree
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// TODO: perform this in a cmd
		m.tracker.reindex(m.tree, m.treeHeight())
	case tui.FilterFocusReqMsg:
		// Focus the filter widget
		blink := m.filter.Focus()
		// Start blinking the cursor.
		return blink
	case tui.FilterBlurMsg:
		// Blur the filter widget
		m.filter.Blur()
		return nil
	case tui.FilterCloseMsg:
		// Close the filter widget
		m.filter.Blur()
		m.filter.SetValue("")
		// Unfilter table items
		return m.buildTree
	case tui.FilterKeyMsg:
		// unwrap key and send to filter widget
		kmsg := tea.KeyMsg(msg)
		var blink tea.Cmd
		m.filter, blink = m.filter.Update(kmsg)
		// Filter table items
		return tea.Batch(
			blink,
			m.buildTree,
		)
	case cursor.BlinkMsg:
		var blink tea.Cmd
		m.filter, blink = m.filter.Update(msg)
		return blink
	}
	return nil
}

func (m model) treeHeight() int {
	// Height of filter widget
	const filterHeight = 2

	if m.filterVisible() {
		return max(0, m.height-filterHeight)
	}
	return m.height
}

func (m model) View() string {
	if m.tree == nil {
		return "building tree"
	}
	var content string
	if m.filterVisible() {
		content = tui.Regular.Margin(0, 1).Render(m.filter.View())
		content += "\n"
		content += strings.Repeat("─", m.width)
		content += "\n"
	}
	treeStyle := lipgloss.NewStyle().
		Width(m.width - tui.ScrollbarWidth).
		MaxWidth(m.width - tui.ScrollbarWidth).
		Inline(true)
	to := lgtree.New().
		Enumerator(enumerator).
		Indenter(indentor)
	m.tree.render(true, to)
	s := to.String()
	lines := strings.Split(s, "\n")
	numVisibleLines := clamp(m.treeHeight(), 0, len(lines))
	visibleLines := lines[m.tracker.start : m.tracker.start+numVisibleLines]
	for i := range visibleLines {
		node := m.tracker.nodes[m.tracker.start+i]
		// Style node according to whether it is the cursor node, selected, or
		// both
		var (
			background lipgloss.Color
			foreground lipgloss.Color
			current    = node == m.tracker.cursorNode
			selected   = m.tracker.isSelected(node)
		)
		if current && selected {
			background = tui.CurrentAndSelectedBackground
			foreground = tui.CurrentAndSelectedForeground
		} else if current {
			background = tui.CurrentBackground
			foreground = tui.CurrentForeground
		} else if selected {
			background = tui.SelectedBackground
			foreground = tui.SelectedForeground
		}
		renderedRow := treeStyle.Render(visibleLines[i])
		// If current row or selected rows, strip colors and apply background color
		if current || selected {
			renderedRow = internal.StripAnsi(renderedRow)
			renderedRow = lipgloss.NewStyle().
				Foreground(foreground).
				Background(background).
				Render(renderedRow)
		}
		visibleLines[i] = renderedRow
	}
	scrollbar := tui.Scrollbar(m.treeHeight(), len(lines), numVisibleLines, m.tracker.start)
	content += lipgloss.JoinHorizontal(lipgloss.Left,
		strings.Join(visibleLines, "\n"),
		scrollbar,
	)
	return content
}

func (m model) BorderText() map[tui.BorderPosition]string {
	modules := fmt.Sprintf(
		"%s%s",
		tui.ModuleIcon(),
		tui.ModuleStyle.Render(fmt.Sprintf("%d", m.tracker.totalModules)),
	)
	workspaces := fmt.Sprintf(
		"%s%s",
		tui.WorkspaceIcon(),
		tui.WorkspaceStyle.Render(fmt.Sprintf("%d", m.tracker.totalWorkspaces)),
	)
	return map[tui.BorderPosition]string{
		tui.BottomMiddle: fmt.Sprintf("%s %s", modules, workspaces),
	}
}

func (m model) buildTree() tea.Msg {
	tree := m.treeBuilder.newTree(m.filter.Value())
	return builtTreeMsg(tree)
}

func (m model) filterVisible() bool {
	// Filter is visible if it's either in focus, or it has a non-empty value.
	return m.filter.Focused() || m.filter.Value() != ""
}

// getWorkspaceIDs retrieves all selected rows, or if no rows are selected, then
// it retrieves the cursor row, and if the rows are workspaces then it returns
// their IDs; if the rows are modules then it returns the IDs of their
// respective *current* workspaces. An error is returned if all modules don't
// have a current workspace or if any other type of rows are selected or are
// currently the cursor row.
func (m model) getWorkspaceIDs() ([]resource.ID, error) {
	kind, ids := m.tracker.getSelectedOrCurrentIDs()
	switch kind {
	case resource.Workspace:
		return ids, nil
	case resource.Module:
		for i, moduleID := range ids {
			mod, err := m.ModuleService.Get(moduleID)
			if err != nil {
				return nil, err
			}
			if mod.CurrentWorkspaceID == nil {
				return nil, errors.New("modules must have a current workspace")
			}
			ids[i] = *mod.CurrentWorkspaceID
		}
		return ids, nil
	default:
		return nil, errors.New("valid only on workspaces and modules")
	}
}

// getModuleIDs retrieves all selected rows, or if no rows are selected, then
// it retrieves the cursor row, and if the rows are workspaces then it returns
// the IDs of their respective parent modules; if the rows are modules then it
// returns their IDs. An error is returned if the rows are not workspaces or
// modules.
func (m model) getModuleIDs() ([]resource.ID, error) {
	kind, ids := m.tracker.getSelectedOrCurrentIDs()
	switch kind {
	case resource.Module:
		return ids, nil
	case resource.Workspace:
		for i, workspaceID := range ids {
			ws, err := m.WorkspaceService.Get(workspaceID)
			if err != nil {
				return nil, err
			}
			ids[i] = ws.ModuleID
		}
		return ids, nil
	default:
		return nil, errors.New("valid only on workspaces and modules")
	}
}
