package explorer

import (
	"path/filepath"
	"slices"
	"strings"

	lgtree "github.com/charmbracelet/lipgloss/tree"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
)

type tree struct {
	value    node
	children []*tree
}

type treeBuilder struct {
	wd               internal.Workdir
	helpers          treeBuilderHelpers
	moduleService    treeBuilderModuleLister
	workspaceService treeBuilderWorkspaceLister
}

type treeBuilderModuleLister interface {
	List() []*module.Module
}

type treeBuilderWorkspaceLister interface {
	List(workspace.ListOptions) []*workspace.Workspace
}

type treeBuilderHelpers interface {
	WorkspaceResourceCount(*workspace.Workspace) string
	WorkspaceCost(ws *workspace.Workspace) string
}

func (b *treeBuilder) newTree(filter string) (*tree, string) {
	t := &tree{
		value: dirNode{root: true, path: b.wd.PrettyString()},
	}
	modules := b.moduleService.List()
	workspaces := b.workspaceService.List(workspace.ListOptions{})
	// Create set of current workspaces for assignment below.
	currentWorkspaces := make(map[resource.ID]bool)
	for _, mod := range modules {
		if mod.CurrentWorkspaceID != nil {
			currentWorkspaces[mod.CurrentWorkspaceID] = true
		}
	}
	// Arrange workspaces by module, for attachment to modules in tree below.
	workspaceNodes := make(map[resource.ID][]workspaceNode, len(modules))
	for _, ws := range workspaces {
		wsNode := workspaceNode{
			id:            ws.ID,
			name:          ws.Name,
			current:       currentWorkspaces[ws.ID],
			resourceCount: b.helpers.WorkspaceResourceCount(ws),
			cost:          b.helpers.WorkspaceCost(ws),
		}
		workspaceNodes[ws.ModuleID] = append(workspaceNodes[ws.ModuleID], wsNode)
	}
	for _, mod := range modules {
		// Set parent to root of tree
		parent := t
		// Split module's path into a list of directories
		for _, dir := range splitDirs(mod.Path) {
			parent = parent.addChild(dirNode{path: dir})
		}
		// The final node is the module tree, with workspaces as children.
		modTree := parent.addChild(moduleNode{
			id:   mod.ID,
			path: mod.Path,
		})
		for _, ws := range workspaceNodes[mod.ID] {
			modTree.addChild(ws)
		}
	}
	// Apply filter if there is one.
	filtered := t.filter(filter)
	// Now render the corresponding lipgloss tree. We do this here rather than
	// in View() because it's quite expensive and thus best kept out of the
	// bubble event loop.
	to := lgtree.New().
		Enumerator(enumerator).
		Indenter(indentor)
	filtered.render(true, to)
	return filtered, to.String()
}

func (t *tree) filter(text string) *tree {
	if text == "" {
		return t
	}
	if strings.Contains(t.value.String(), text) {
		return t
	}
	to := &tree{value: t.value}
	for _, child := range t.children {
		result := child.filter(text)
		if result != nil {
			to.children = append(to.children, result)
		}
	}
	if len(to.children) == 0 {
		if dir, ok := to.value.(dirNode); ok {
			if dir.root {
				// If nothing matches return at least the root node.
				return to
			}
		}
		return nil
	}
	return to
}

func (t *tree) render(root bool, to *lgtree.Tree) {
	s := t.value.String()
	lgnode := lgtree.Root(s)
	// First node in tracker is the root node.
	if root {
		to.Root(lgnode)
		lgnode = to
	} else {
		to.Child(lgnode)
	}
	for _, child := range t.children {
		child.render(false, lgnode)
	}
}

// addChild adds a child to the tree; if child is already in tree then no action
// is taken and the existing child is returned. If the child is added, the
// children are sorted and the new child is returned.
func (t *tree) addChild(child node) *tree {
	for _, existing := range t.children {
		if existing.value == child {
			return existing
		}
	}
	newTree := &tree{value: child}
	t.children = append(t.children, newTree)
	// keep children ordered
	slices.SortFunc(t.children, func(a, b *tree) int {
		// directories come before modules
		if _, ok := a.value.(dirNode); ok {
			if _, ok := b.value.(moduleNode); ok {
				return -1
			}
		}
		if _, ok := a.value.(moduleNode); ok {
			if _, ok := b.value.(dirNode); ok {
				return 1
			}
		}
		// otherwise order lexicographically.
		if a.value.Value() < b.value.Value() {
			return -1
		}
		return 1
	})
	return newTree
}

func splitDirs(path string) []string {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	parts := strings.Split(dir, string(filepath.Separator))
	dirs := make([]string, len(parts))
	for i, p := range parts {
		if i > 0 {
			dirs[i] = filepath.Join(dirs[i-1], p)
		} else {
			dirs[i] = p
		}
	}
	return dirs
}

func indentor(children lgtree.Children, index int) string {
	if children.Length()-1 == index {
		return " "
	}
	return "│"
}

func enumerator(children lgtree.Children, index int) string {
	if children.Length()-1 == index {
		return "└"
	}
	return "├"
}
