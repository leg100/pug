package explorer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lgtree "github.com/charmbracelet/lipgloss/tree"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/workspace"
)

type Maker struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Helpers    *tui.Helpers
	Workdir    internal.Workdir
}

func (m *Maker) Make(_ resource.ID, width, height int) (tea.Model, error) {
	return model{
		WorkspaceService: m.Workspaces,
		ModuleService:    m.Modules,
		Helpers:          m.Helpers,
		Workdir:          m.Workdir,
		tree:             newTree(m.Workdir, nil, nil),
		tracker: &tracker{
			selectedNodes: make(map[fmt.Stringer]int),
		},
		w: width,
		h: height,
	}, nil
}

type model struct {
	*tui.Helpers

	ModuleService    *module.Service
	WorkspaceService *workspace.Service
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
			m.tracker.deselectAll()
			return m, nil
		case key.Matches(msg, keys.Global.SelectRange):
			err := m.tracker.selectRange()
			return m, tui.ReportError(err)
		case key.Matches(msg, localKeys.Enter):
			m.tracker.toggleClose()
			return m, nil
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
	// Create map of selected node indices for faster lookup
	selectedLineIndices := make(map[int]struct{})
	for _, i := range m.tracker.selectedNodes {
		selectedLineIndices[i] = struct{}{}
	}
	for i := range lines {
		style := lipgloss.NewStyle().
			Width(m.w - tui.ScrollbarWidth)
		trackerIndex := m.tracker.start + i
		// Style node according to whether it is the cursor node, selected, or
		// both
		if trackerIndex == m.tracker.cursorIndex {
			if _, ok := selectedLineIndices[trackerIndex]; ok {
				style = style.
					Background(tui.CurrentAndSelectedBackground).
					Foreground(tui.CurrentAndSelectedForeground)
			} else {
				style = style.
					Background(tui.CurrentBackground).
					Foreground(tui.CurrentForeground)
			}
		} else if _, ok := selectedLineIndices[trackerIndex]; ok {
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
	return fmt.Sprintf("start: %d", m.tracker.start)
}

func (m model) buildTree() tea.Msg {
	tree := newTree(m.Workdir, m.modules, m.workspaces)
	return builtTreeMsg(tree)
}
