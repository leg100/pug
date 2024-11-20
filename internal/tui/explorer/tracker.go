package explorer

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/tui"
)

// tracker tracks the cursor node and any selected nodes
type tracker struct {
	nodes []fmt.Stringer

	cursorNode fmt.Stringer
	cursorPos  int
}

func (t *tracker) reindex(tree *tree) {
	t.nodes = nil
	t.traverse(tree)

	if t.cursorNode == nil && len(t.nodes) > 0 {
		t.cursorNode = t.nodes[0]
		t.cursorPos = 0
	}
}

func (t *tracker) traverse(tree *tree) {
	if tree.value == t.cursorNode {
		t.cursorPos = len(t.nodes) - 1
	}
	t.nodes = append(t.nodes, tree.value)
	for _, child := range tree.children {
		t.traverse(child)
	}
}

func (t *tracker) renderedCursorNode() string {
	return lipgloss.NewStyle().
		Background(tui.CurrentBackground).
		Foreground(tui.CurrentForeground).
		Render(t.cursorNode.String())
}

func (t *tracker) cursorUp() {
	t.cursorPos = max(t.cursorPos-1, 0)
	t.cursorNode = t.nodes[t.cursorPos]
}

func (t *tracker) cursorDown() {
	t.cursorPos = min(t.cursorPos+1, len(t.nodes)-1)
	t.cursorNode = t.nodes[t.cursorPos]
}
