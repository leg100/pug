package explorer

import (
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/maps"
)

// tracker tracks the cursor node and any selected nodes, as well as which nodes
// are currently visible.
type tracker struct {
	nodes       []node
	cursorNode  node
	cursorIndex int
	// index of first visible row
	start int
	// total number of modules
	totalModules int
	// total number of workspaces
	totalWorkspaces int

	*selector
}

func newTracker(tree *tree, height int) *tracker {
	t := &tracker{
		selector: &selector{
			selections: make(map[resource.ID]struct{}),
		},
	}
	t.reindex(tree, height)
	return t
}

func (t *tracker) reindex(tree *tree, height int) {
	t.nodes = nil
	t.totalModules = 0
	t.totalWorkspaces = 0
	t.doReindex(tree)

	if t.cursorNode == nil && len(t.nodes) > 0 {
		t.cursorNode = t.nodes[0]
		t.cursorIndex = 0
	}
	t.setStart(height)
}

func (t *tracker) doReindex(tree *tree) {
	t.nodes = append(t.nodes, tree.value)
	// maintain tally of numbers of types of nodes
	switch tree.value.(type) {
	case moduleNode:
		t.totalModules++
	case workspaceNode:
		t.totalWorkspaces++
	}
	// Track index of cursor node
	if t.cursorNode != nil && t.cursorNode.ID() == tree.value.ID() {
		t.cursorIndex = len(t.nodes) - 1
	}
	// Track indices of selected nodes
	if dir, ok := tree.value.(dirNode); ok {
		if dir.closed {
			return
		}
	}
	for _, child := range tree.children {
		t.doReindex(child)
	}
}

func (t *tracker) moveCursor(delta, treeHeight int) {
	t.cursorIndex = clamp(t.cursorIndex+delta, 0, len(t.nodes)-1)
	if len(t.nodes) > 0 {
		t.cursorNode = t.nodes[t.cursorIndex]
	}
	t.setStart(treeHeight)
}

func (t *tracker) toggleSelection() error {
	if t.cursorNode == nil {
		return nil
	}
	return t.selector.toggle(t.cursorNode)
}

func (t *tracker) selectAll() error {
	if t.cursorNode == nil {
		return nil
	}
	return t.selector.addAll(t.cursorNode, t.nodes...)
}

// selectRange selects a range of nodes. If th cursor node is after a selected
// node then the rows between them are selected, including the cursor node.
// Otherwise, if the cursor node is before a selected node then nodes between
// them are selected, including the cursor node. If there are no selected nodes
// then no action is taken.
func (t *tracker) selectRange() error {
	if t.cursorNode == nil {
		return nil
	}
	return t.selector.addRange(t.cursorNode, t.cursorIndex, t.nodes...)
}

func (t *tracker) getSelectedOrCurrentIDs() (resource.Kind, []resource.ID) {
	if len(t.selections) == 0 {
		id, ok := t.cursorNode.ID().(resource.ID)
		if !ok {
			// TODO: consider returning error
			return -1, nil
		}
		return id.Kind, []resource.ID{id}
	}
	return *t.selector.kind, maps.Keys(t.selections)
}

func (t *tracker) getCursorID() *resource.ID {
	id, ok := t.cursorNode.ID().(resource.ID)
	if !ok {
		// TODO: consider returning error
		return nil
	}
	return &id
}

func (t *tracker) toggleClose() {
	if t.cursorNode == nil {
		return
	}
	if dir, ok := t.cursorNode.(dirNode); ok {
		dir.closed = !dir.closed
	}
}

func (t *tracker) setStart(height int) {
	// Start index must be at least the cursor position minus the max number
	// of visible nodes.
	minimum := max(0, t.cursorIndex-height+1)
	// Start index must be at most the lesser of:
	// (a) the cursor position, or
	// (b) the number of nodes minus the maximum number of visible rows (as many
	// rows as possible are rendered)
	maximum := max(0, min(t.cursorIndex, len(t.nodes)-height))
	t.start = clamp(t.start, minimum, maximum)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
