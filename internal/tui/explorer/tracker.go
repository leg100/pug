package explorer

import (
	"fmt"
)

// tracker tracks the cursor node and any selected nodes
type tracker struct {
	nodes []fmt.Stringer

	cursorNode fmt.Stringer
	cursorPos  int

	selectedNodes map[fmt.Stringer]struct{}
}

func (t *tracker) reindex(tree *tree) {
	t.nodes = nil
	t.traverse(tree.root)

	if t.cursorNode == nil && len(t.nodes) > 0 {
		t.cursorNode = t.nodes[0]
		t.cursorPos = 0
	}
}

func (t *tracker) traverse(n *node) {
	if n.value == t.cursorNode {
		t.cursorPos = len(t.nodes) - 1
	}
	t.nodes = append(t.nodes, n.value)
	for _, child := range n.children {
		t.traverse(child)
	}
}

func (t *tracker) cursorUp() {
	t.cursorPos = max(t.cursorPos-1, 0)
	t.cursorNode = t.nodes[t.cursorPos]
}

func (t *tracker) cursorDown() {
	t.cursorPos = min(t.cursorPos+1, len(t.nodes)-1)
	t.cursorNode = t.nodes[t.cursorPos]
}

func (t *tracker) toggleSelection() {
	if t.cursorNode == nil {
		return
	}
	if _, ok := t.selectedNodes[t.cursorNode]; ok {
		// de-select
		delete(t.selectedNodes, t.cursorNode)
	} else {
		// select
		t.selectedNodes[t.cursorNode] = struct{}{}
	}
}
