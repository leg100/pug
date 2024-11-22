package explorer

import (
	"fmt"
)

// tracker tracks the cursor node and any selected nodes, as well as which nodes
// are currently visible.
type tracker struct {
	nodes       []fmt.Stringer
	cursorNode  fmt.Stringer
	cursorIndex int
	// selectedNodes maps nodes that are currently selected to their positional
	// index.
	selectedNodes map[fmt.Stringer]int
	// index of first visible row
	start int
	// height of tree widget
	height int
}

func (t *tracker) reindex(tree *tree) {
	t.nodes = nil
	t.doReindex(tree)

	if t.cursorNode == nil && len(t.nodes) > 0 {
		t.cursorNode = t.nodes[0]
		t.cursorIndex = 0
	}
	t.setStart()
}

func (t *tracker) doReindex(tree *tree) {
	t.nodes = append(t.nodes, tree.value)
	if tree.value == t.cursorNode {
		t.cursorIndex = len(t.nodes) - 1
	}
	for _, child := range tree.children {
		t.doReindex(child)
	}
}

func (t *tracker) cursorUp() {
	t.cursorIndex = max(t.cursorIndex-1, 0)
	if len(t.nodes) > 0 {
		t.cursorNode = t.nodes[t.cursorIndex]
	}
	t.setStart()
}

func (t *tracker) cursorDown() {
	t.cursorIndex = min(t.cursorIndex+1, len(t.nodes)-1)
	if len(t.nodes) > 0 {
		t.cursorNode = t.nodes[t.cursorIndex]
	}
	t.setStart()
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
		t.selectedNodes[t.cursorNode] = t.cursorIndex
	}
}

func (t *tracker) setStart() {
	// Start index must be at least the cursor position minus the max number
	// of visible nodes.
	minimum := max(0, t.cursorIndex-t.height+1)
	// Start index must be at most the lesser of:
	// (a) the cursor position, or
	// (b) the number of nodes minus the maximum number of visible rows (as many
	// rows as possible are rendered)
	maximum := max(0, min(t.cursorIndex, len(t.nodes)-t.height))
	t.start = clamp(t.start, minimum, maximum)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
