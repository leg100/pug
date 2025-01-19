package explorer

import (
	"errors"

	"github.com/leg100/pug/internal/resource"
)

var (
	ErrSelectionKindMismatch = errors.New("selections must be of the same kind")
	ErrUnselectableNode      = errors.New("node is not selectable")
)

// selector manages the selection of nodes. Only resources may be selected (i.e.
// not directories), and resources must be of the same kind.
// selected.
type selector struct {
	selections map[resource.ID]struct{}
	kind       *resource.Kind
}

func (s *selector) add(n node) error {
	id, ok := n.ID().(resource.MonotonicID)
	if !ok {
		return ErrUnselectableNode
	}
	// Error if node type is different to existing selected node type.
	if s.kind != nil {
		if *s.kind != id.Kind {
			return ErrSelectionKindMismatch
		}
	} else {
		s.kind = &id.Kind
	}
	s.selections[id] = struct{}{}
	return nil
}

func (s *selector) reindex(nodes []node) {
	if len(s.selections) == 0 {
		return
	}
	selections := make(map[resource.ID]struct{}, len(s.selections))
	for _, n := range nodes {
		id, ok := n.ID().(resource.MonotonicID)
		if !ok {
			continue
		}
		if _, ok := s.selections[id]; ok {
			selections[id] = struct{}{}
		}
	}
	s.selections = selections
}

// addAll implements "select all" functionality with a twist: it ensures only
// nodes of the same type of selected; the cursor node must match any existing
// selection type, and then the nodes are filtered to only add those
// matching the cursor type.
func (s *selector) addAll(cursor node, nodes ...node) error {
	if err := s.add(cursor); err != nil {
		return err
	}
	for _, node := range nodes {
		if err := s.add(node); err != nil {
			// filter out unselectable/incompatible nodes
			continue
		}
	}
	return nil
}

// addRange implements "select range" functionality: if the cursor node is after
// a selected node then the nodes between are selected, including the cursor
// node. Otherwise, if the cursor node is *before* a selected node then nodes
// between are selected, including the cursor node. If there are no selected
// nodes no action is taken.
//
// The cursor node and the existing selection must be of the same type. If nodes
// between the cursor and the existing selection are of a different type then
// they are skipped.
func (s *selector) addRange(cursor node, cursorIndex int, nodes ...node) error {
	if len(s.selections) == 0 {
		return nil
	}
	if err := s.add(cursor); err != nil {
		return err
	}
	// Determine first node to select, and the number of nodes to select.
	var (
		first = -1
		n     = 0
	)
	for i, node := range nodes {
		if i == cursorIndex && first > -1 && first < cursorIndex {
			// Select nodes before cursor node
			n = cursorIndex - first
			break
		}
		if !s.isSelected(node) {
			// Ignore unselected nodes
			continue
		}
		if i > cursorIndex {
			// Select nodes after cursor node and all nodes up to but not
			// including next selected node
			first = cursorIndex + 1
			n = i - cursorIndex
			break
		}
		// Start selecting nodes after this currently selected node.
		first = i + 1
	}
	for _, node := range nodes[first : first+n] {
		if err := s.add(node); err != nil {
			// skip incompatible nodes
			continue
		}
	}
	return nil
}

func (s *selector) remove(n node) {
	id, ok := n.ID().(resource.MonotonicID)
	if !ok {
		// silently ignore request to remove non-resource node
		return
	}
	delete(s.selections, id)
	if len(s.selections) == 0 {
		s.kind = nil
	}
}

func (s *selector) removeAll() {
	s.selections = make(map[resource.ID]struct{})
	s.kind = nil
}

func (s *selector) toggle(n node) error {
	if s.isSelected(n) {
		s.remove(n)
		return nil
	}
	return s.add(n)
}

func (s *selector) isSelected(n node) bool {
	id, ok := n.ID().(resource.MonotonicID)
	if !ok {
		// non-resource nodes cannot be selected
		return false
	}
	_, ok = s.selections[id]
	return ok
}
