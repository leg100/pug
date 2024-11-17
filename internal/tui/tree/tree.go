package tree

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/tree"
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
		Workspaces: m.Workspaces,
		Modules:    m.Modules,
		Helpers:    m.Helpers,
		Workdir:    m.Workdir,
	}, nil
}

type model struct {
	*tui.Helpers

	Modules    *module.Service
	Workspaces *workspace.Service
	Workdir    internal.Workdir
}

func (m model) Init() tea.Cmd {
	return m.buildTree()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) View() string {
	return ""
}

func (m *model) buildTree() tea.Cmd {
	// Retrieve all modules and workspaces
	modules := m.Modules.List()
	workspaces := m.Workspaces.List(workspace.ListOptions{})

	// Create new tree rooted at working directory.
	t := tree.Root(dir{name: m.Workdir.String()})
	for _, module := range modules {
		// Set parent to root of tree
		parent := t
		// Split its path into a list of directories
		dirs := strings.Split(module.Path, string(filepath.Separator))
		// Iterate over each directory.
		for j, d := range dirs {
			// The final directory is the module node
			if j == len(dirs)-1 {
				// Add module as child to parent tree
				parent.Child(mod{name: d})
				break
			} else {
				// Check if directory has already been added to last tree
				node := dir{name: d}
				// Keep track of previous child
				var i int
				// Add children of previous tree one-by-one to parent.
				for ; i < parent.Children().Length(); i++ {
					if node.String() == parent.Children().At(i).String() {
						// Already added, so nothing to do.
						reak
					} else if node.String() > parent.Children().At(i).String() {
						// Previous child comes before current node, so add
						// previous child to parent
						parent.Child(lastChild)
					} else {
						// Previous child comes after current node, so add
						// current node to parent, and break.
						parent.Child(node)
						break
					}
				}
				// Add remaining children from last tree to parent. We have to
				// do this one-by-one because *tree.Child() expects a slice of
				// any.
				for k := i; k < last.Children().Length(); k++ {
					parent.Child(last.Children().At(k))
				}
				// Set
			}
			// Set parent to be the current node
		}

		// pass each child tree successively to the module tree, which uses
		// auto-nesting to automatically parent each child tree to its sibling.
		root.Child(childTrees...)
	}
}

func mergeTrees(a, b *tree.Tree) *tree.Tree {
}
