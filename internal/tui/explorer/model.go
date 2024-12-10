package explorer

import (
	"errors"
	"fmt"
	"strings"

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

// Height of filter widget
const filterHeight = 2

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
			m.tracker.cursorUp(m.treeHeight())
			return nil
		case key.Matches(msg, keys.Navigation.LineDown):
			m.tracker.cursorDown(m.treeHeight())
			return nil
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
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Module {
				return tui.ReportError(errors.New("can only trigger init on modules"))
			}
			fn := func(moduleID resource.ID) (task.Spec, error) {
				return m.Modules.Init(moduleID, upgrade)
			}
			return m.CreateTasks(fn, ids...)
		case key.Matches(msg, keys.Common.Validate):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Module {
				return tui.ReportError(errors.New("can only trigger init on modules"))
			}
			cmd := m.CreateTasks(m.Modules.Validate, ids...)
			return cmd
		case key.Matches(msg, keys.Common.Format):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Module {
				return tui.ReportError(errors.New("can only trigger format on modules"))
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
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Workspace {
				return tui.ReportError(errors.New("can only trigger plans on workspaces"))
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
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Workspace {
				return tui.ReportError(errors.New("can only trigger applies on workspaces"))
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Apply(workspaceID, createPlanOptions)
			}
			return tui.YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(ids)),
				m.CreateTasks(fn, ids...),
			)
		case key.Matches(msg, keys.Common.Cost):
			kind, ids := m.tracker.getSelectedOrCurrentIDs()
			if kind != resource.Workspace {
				return tui.ReportError(errors.New("can only trigger infracost on workspaces"))
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
	}
	return nil
}

func (m model) treeHeight() int {
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
	// MaxHeight(m.height).
	return content
}

func (m model) Metadata() string {
	e := lipgloss.NewStyle().
		Foreground(tui.DarkRed).
		Render("e")
	return fmt.Sprintf("%sxplorer", e)
}

func (m model) buildTree() tea.Msg {
	tree := m.treeBuilder.newTree(m.filter.Value())
	return builtTreeMsg(tree)
}

func (m model) filterVisible() bool {
	// Filter is visible if it's either in focus, or it has a non-empty value.
	return m.filter.Focused() || m.filter.Value() != ""
}
