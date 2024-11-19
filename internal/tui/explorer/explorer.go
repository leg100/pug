package explorer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	case initMsg:
		m.modules = msg.modules
		m.workspaces = msg.workspaces
		return m, m.buildTree
	case builtTreeMsg:
		m.tree = (*tree)(msg)
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
	}
	return m, nil
}

func (m model) View() string {
	if m.tree == nil {
		return "building tree"
	}
	return lipgloss.NewStyle().
		Height(m.h).
		MaxHeight(m.h).
		Render(m.tree.Render())
}

func (m model) buildTree() tea.Msg {
	tree := newTree(m.Workdir, m.modules, m.workspaces)
	return builtTreeMsg(tree)
}
