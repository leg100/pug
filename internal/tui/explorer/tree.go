package explorer

import (
	"fmt"
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
	value    fmt.Stringer
	children []*tree
}

func newTree(wd internal.Workdir, modules []*module.Module, workspaces []*workspace.Workspace) *tree {
	t := &tree{
		value: dirNode{root: true, path: wd.PrettyString()},
	}
	// Arrange workspaces by module, for attachment to modules in tree below.
	workspaceNodes := make(map[resource.ID][]workspaceNode, len(modules))
	for _, ws := range workspaces {
		workspaceNodes[ws.ModuleID] = append(workspaceNodes[ws.ModuleID], workspaceNode{
			id:   ws.ID,
			name: ws.Name,
		})
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
	return t
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
func (t *tree) addChild(child fmt.Stringer) *tree {
	for _, existing := range t.children {
		if existing.value == child {
			return existing
		}
	}
	newTree := &tree{value: child}
	t.children = append(t.children, newTree)
	// keep children lexicographically ordered
	slices.SortFunc(t.children, func(a, b *tree) int {
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
