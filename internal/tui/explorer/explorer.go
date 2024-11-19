package explorer

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/explorer/tree"
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

	modules      []*module.Module
	workspaces   []*workspace.Workspace
	renderedTree string
	w, h         int
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
	case renderedTreeMsg:
		m.renderedTree = string(msg)
		return m, nil
	case tea.WindowSizeMsg:
		m.w = msg.Width
		m.h = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	return lipgloss.NewStyle().
		MaxHeight(m.h).
		Render(m.renderedTree)
}

func (m *model) buildTree() tea.Msg {
	dirWithIcon := func(dir string) string {
		return fmt.Sprintf("%s %s", dirIcon, dir)
	}
	// Arrange workspaces by module, for attachment to modules in tree below.
	workspacesByModuleID := make(map[resource.ID][]any, len(m.modules))
	for _, ws := range m.workspaces {
		workspaceWithIcon := fmt.Sprintf("%s %s", workspaceIcon, ws.Name)
		workspacesByModuleID[ws.ModuleID] = append(workspacesByModuleID[ws.ModuleID], workspaceWithIcon)
	}
	// Build UI tree from data tree
	t := tree.Root(dirWithIcon(m.Workdir.String())).
		Indenter(indentor).
		Enumerator(enumerator)
	for _, mod := range m.modules {
		// Set parent to root of tree
		parent := t
		// Split its path into a list of directories
		dirs := strings.Split(mod.Path, string(filepath.Separator))
		// Iterate over each directory.
		for _, d := range dirs {
			// Insert dir node into tree if not already added
			node := tree.Root(dirWithIcon(d))
			var added bool
			for i := range parent.Children().Length() {
				if node.Value() == parent.Children().At(i).Value() {
					// Already added, no changes necessary to children, make
					// node the new parent.
					node = parent.Children().At(i).(*tree.Tree)
					added = true
					break
				} else if node.Value() < parent.Children().At(i).Value() {
					// Insert node according to lexicographic order.
					parent.InsertChild(node, i)
					added = true
					break
				}
			}
			if !added {
				// Node not added so add it now.
				parent.Child(node)
			}
			// Set node to be the new parent
			parent = node
		}
		// The final node is the module tree, with workspaces as children.
		moduleWorkspaces := workspacesByModuleID[mod.ID]
		modTree := tree.Root(mod).Child(moduleWorkspaces...)
		// Add module tree to parent.
		parent.Child(modTree)
	}
	return renderedTreeMsg(t.String())
}

func indentor(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return " "
	}
	return "│"
}

func enumerator(children tree.Children, index int) string {
	if children.Length()-1 == index {
		return "└"
	}
	return "├"
}
