package explorer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lgtree "github.com/charmbracelet/lipgloss/tree"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
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
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.tracker.cursorUp()
			return m, nil
		case "down", "j":
			m.tracker.cursorDown()
			return m, nil
		case " ":
			m.tracker.toggleSelection()
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
	for i := range lines {
		style := lipgloss.NewStyle().
			Width(m.w - tui.ScrollbarWidth)
		// Style node if cursor is on node
		if m.tracker.start+i == m.tracker.cursorIndex {
			style = style.
				Background(tui.CurrentBackground).
				Foreground(tui.CurrentForeground)
		}
		// Style node if selected
		for _, pos := range m.tracker.selectedNodes {
			if m.tracker.start+i == pos {
				style = style.
					Background(tui.SelectedBackground).
					Foreground(tui.SelectedForeground)
			}
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
