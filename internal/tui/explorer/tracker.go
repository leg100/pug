package explorer

import (
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/exp/maps"
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
	// Track index of cursor node
	if tree.value == t.cursorNode {
		t.cursorIndex = len(t.nodes) - 1
	}
	if dir, ok := tree.value.(dirNode); ok {
		if dir.closed {
			return
		}
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

func (t *tracker) toggleSelection() error {
	if t.cursorNode == nil {
		return nil
	}
	if _, ok := t.selectedNodes[t.cursorNode]; ok {
		// de-select
		delete(t.selectedNodes, t.cursorNode)
	} else {
		// select - user can only select nodes of the same type, i.e. all
		// modules, or all workspaces, so if there is at least one existing
		// selection then check its type is equal to the cursor node.
		if len(t.selectedNodes) > 0 {
			existingSelection := maps.Keys(t.selectedNodes)[0]
			if reflect.TypeOf(existingSelection) != reflect.TypeOf(t.cursorNode) {
				return errors.New("selections must be of the same type")
			}
		}
		t.selectedNodes[t.cursorNode] = t.cursorIndex
	}
	return nil
}

func (t *tracker) selectAll() error {
	if t.cursorNode == nil {
		return nil
	}
	// User can only select nodes of the same type, i.e. all
	// modules, or all workspaces.
	// Retrieve type of cursor node and error if any existing selections are of
	// a different type.
	cursorType := reflect.TypeOf(t.cursorNode)
	if len(t.selectedNodes) > 0 {
		existingSelection := maps.Keys(t.selectedNodes)[0]
		if reflect.TypeOf(existingSelection) != cursorType {
			return errors.New("selections must be of the same type")
		}
	}
	for i, node := range t.nodes {
		if reflect.TypeOf(node) != cursorType {
			// Skip nodes of a different type
			continue
		}
		t.selectedNodes[node] = i
	}
	return nil
}

func (t *tracker) deselectAll() {
	t.selectedNodes = make(map[fmt.Stringer]int)
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
	if len(t.selectedNodes) == 0 {
		return nil
	}
	// Error if cursor node type is different to existing selected node type.
	cursorType := reflect.TypeOf(t.cursorNode)
	existingSelection := maps.Keys(t.selectedNodes)[0]
	if reflect.TypeOf(existingSelection) != cursorType {
		return errors.New("selections must be of the same type")
	}
	// Determine first node to select, and the number of nodes to select.
	var (
		first = -1
		n     = 0
	)
	for i, node := range t.nodes {
		if i == t.cursorIndex && first > -1 && first < t.cursorIndex {
			// Select nodes before and including cursor node
			n = t.cursorIndex - first + 1
			break
		}
		if _, ok := t.selectedNodes[node]; !ok {
			// Ignore unselected nodes
			continue
		}
		if i > t.cursorIndex {
			// Select nodes including cursor node and all nodes up to but not
			// including next selected node
			first = t.cursorIndex
			n = i - t.cursorIndex
			break
		}
		// Start selecting nodes after this currently selected node.
		first = i + 1
	}
	for i, node := range t.nodes[first : first+n] {
		if reflect.TypeOf(node) != cursorType {
			// Skip nodes of a different type
			continue
		}
		t.selectedNodes[node] = first + i
	}
	return nil
}

func (t *tracker) toggleClose() {
	if t.cursorNode == nil {
		return
	}
	if dir, ok := t.cursorNode.(dirNode); ok {
		dir.closed = !dir.closed
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
