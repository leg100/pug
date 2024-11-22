package explorer

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	lgtree "github.com/charmbracelet/lipgloss/tree"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/workspace"
)

type tree struct {
	*tracker

	root *node
}

type node struct {
	value    fmt.Stringer
	children []*node
}

func newTree(wd internal.Workdir, modules []*module.Module, workspaces []*workspace.Workspace) *tree {
	// Arrange workspaces by module, for attachment to modules in tree below.
	workspaceNodes := make(map[resource.ID][]workspaceNode, len(modules))
	for _, ws := range workspaces {
		workspaceNodes[ws.ModuleID] = append(workspaceNodes[ws.ModuleID], workspaceNode{
			id:   ws.ID,
			name: ws.Name,
		})
	}
	t := &tree{
		root: &node{
			value: dirNode{root: true, path: wd.PrettyString()},
		},
		tracker: &tracker{
			selectedNodes: make(map[fmt.Stringer]struct{}),
		},
	}
	for _, mod := range modules {
		// Set parent to root of tree
		parent := t.root
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
	t.tracker.reindex(t)
	return t
}

func (t *tree) render() string {
	to := lgtree.New().
		Enumerator(enumerator).
		Indenter(indentor)
	t.renderNode(t.root, to)
	return to.String()
}

func (t *tree) renderNode(from *node, to *lgtree.Tree) {
	s := from.value.String()
	// Style node if cursor is on node
	if t.tracker.cursorNode == from.value {
		s = lipgloss.NewStyle().
			Background(tui.CurrentBackground).
			Foreground(tui.CurrentForeground).
			Render(s)
	}
	// Style node if selected
	if _, ok := t.tracker.selectedNodes[from.value]; ok {
		s = lipgloss.NewStyle().
			Background(tui.SelectedBackground).
			Foreground(tui.SelectedForeground).
			Render(s)
	}
	lgnode := lgtree.Root(s)
	// First node in tracker is the root node.
	if t.tracker.nodes[0] == from.value {
		to.Root(lgnode)
		lgnode = to
	} else {
		to.Child(lgnode)
	}
	for _, child := range from.children {
		t.renderNode(child, lgnode)
	}
}

// addChild adds a child to the tree; if child is already in tree then no action
// is taken and the existing child is returned. If the child is added, the
// children are sorted and the new child is returned.
func (n *node) addChild(child fmt.Stringer) *node {
	for _, existing := range n.children {
		if existing.value == child {
			return existing
		}
	}
	newTree := &node{value: child}
	n.children = append(n.children, newTree)
	// keep children lexicographically ordered
	slices.SortFunc(n.children, func(a, b *node) int {
		if a.value.String() < b.value.String() {
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
