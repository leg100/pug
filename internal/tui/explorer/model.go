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
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/workspace"
)

type Maker struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	Workdir          internal.Workdir
	Helpers          *tui.Helpers
}

func (mm *Maker) Make(id resource.ID, width, height int) (tui.ChildModel, error) {
	builder := &treeBuilder{
		wd:               mm.Workdir,
		helpers:          mm.Helpers,
		moduleService:    mm.ModuleService,
		workspaceService: mm.WorkspaceService,
	}
	filter := textinput.New()
	filter.Prompt = "Filter: "
	tree, lgtree := builder.newTree("")
	m := &model{
		Helpers:      mm.Helpers,
		Workdir:      mm.Workdir,
		treeBuilder:  builder,
		tree:         tree,
		renderedTree: lgtree,
		tracker:      newTracker(tree, height),
		filter:       filter,
	}
	m.common = &tui.ActionHandler{
		Helpers:     mm.Helpers,
		IDRetriever: m,
	}
	return m, nil
}

type model struct {
	*tui.Helpers

	Workdir internal.Workdir

	common        *tui.ActionHandler
	treeBuilder   *treeBuilder
	tree          *tree
	tracker       *tracker
	renderedTree  string
	width, height int
	filter        textinput.Model
	status        buildTreeStatus
}

type buildTreeStatus int

const (
	notBuildingTree buildTreeStatus = iota
	buildingTree
	queueBuildTree
)

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.buildTree,
		reload(true, m.Modules),
	)
}

func (m *model) Update(msg tea.Msg) tea.Cmd {
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
		case key.Matches(msg, localKeys.SetCurrentWorkspace):
			ws, ok := m.tracker.cursorNode.(workspaceNode)
			if !ok {
				return tui.ReportError(errors.New("cursor is not on a workspace"))
			}
			return func() tea.Msg {
				if err := m.Workspaces.SelectWorkspace(ws.id); err != nil {
					return fmt.Errorf("setting current workspace: %w", err)
				}
				return tui.InfoMsg("set current workspace to " + ws.name)
			}
		case key.Matches(msg, keys.Common.Delete):
			ws, ok := m.tracker.cursorNode.(workspaceNode)
			if !ok {
				return tui.ReportError(errors.New("cursor is not on a workspace"))
			}
			return tui.YesNoPrompt(
				fmt.Sprintf("Delete workspace %s?", ws.name),
				m.CreateTasks(m.Workspaces.Delete, ws.id),
			)
		case key.Matches(msg, localKeys.ReloadWorkspaces):
			ids, err := m.GetModuleIDs()
			if err != nil {
				return tui.ReportError(err)
			}
			return m.CreateTasks(m.Workspaces.Reload, ids...)
		case key.Matches(msg, localKeys.ReloadModules):
			return reload(false, m.Modules)
		default:
			return m.common.Update(msg)
		}
	case builtTreeMsg:
		m.tree = msg.tree
		m.renderedTree = msg.rendered
		// TODO: perform this in a cmd
		m.tracker.reindex(m.tree, m.treeHeight())
		if m.status == queueBuildTree {
			return m.buildTree
		} else {
			m.status = notBuildingTree
		}
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
	lines := strings.Split(m.renderedTree, "\n")
	numVisibleLines := clamp(m.treeHeight(), 0, len(lines))
	visibleLines := lines[m.tracker.start : m.tracker.start+numVisibleLines]
	for i := range visibleLines {
		node := m.tracker.nodes[m.tracker.start+i]
		// Style node according to whether it is the cursor node, selected, or
		// both
		var (
			background lipgloss.Color
			foreground lipgloss.Color
			current    = node.ID() == m.tracker.cursorNode.ID()
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
		tui.TopMiddleBorder:    "PUG 󰩃 ",
		tui.BottomMiddleBorder: fmt.Sprintf("%s %s", modules, workspaces),
	}
}

func (m *model) buildTree() tea.Msg {
	switch m.status {
	case notBuildingTree:
		tree, rendered := m.treeBuilder.newTree(m.filter.Value())
		return builtTreeMsg{
			tree:     tree,
			rendered: rendered,
		}
	case buildingTree:
		m.status = queueBuildTree
	}
	return nil
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
func (m model) GetWorkspaceIDs() ([]resource.ID, error) {
	kind, ids := m.tracker.getSelectedOrCurrentIDs()
	switch kind {
	case resource.Workspace:
		return ids, nil
	case resource.Module:
		for i, moduleID := range ids {
			mod, err := m.Modules.Get(moduleID)
			if err != nil {
				return nil, err
			}
			if mod.CurrentWorkspaceID == nil {
				return nil, errors.New("modules must have a current workspace")
			}
			ids[i] = mod.CurrentWorkspaceID
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
func (m model) GetModuleIDs() ([]resource.ID, error) {
	kind, ids := m.tracker.getSelectedOrCurrentIDs()
	switch kind {
	case resource.Module:
		return ids, nil
	case resource.Workspace:
		for i, workspaceID := range ids {
			ws, err := m.Workspaces.Get(workspaceID)
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

func (m model) HelpBindings() []key.Binding {
	bindings := m.common.HelpBindings()
	// Only show these help bindings when the cursor is on a workspace.
	if _, ok := m.tracker.cursorNode.(workspaceNode); ok {
		bindings = append(bindings, localKeys.SetCurrentWorkspace)
		bindings = append(bindings, keys.Common.Delete)
	}
	return bindings
}
