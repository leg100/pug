package explorer

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

type node interface {
	fmt.Stringer
	// ID uniquely identifies the node
	ID() any
	// Value returns a value for lexicographic sorting
	Value() string
}

type dirNode struct {
	path string
	root bool
}

func (d dirNode) ID() any {
	return d.path
}

func (d dirNode) Value() string {
	if d.root {
		return d.path
	} else {
		return filepath.Base(d.path)
	}
}

func (d dirNode) String() string {
	return fmt.Sprintf("%s %s", tui.DirIcon, d.Value())
}

type moduleNode struct {
	id   resource.MonotonicID
	path string
}

func (m moduleNode) ID() any {
	return m.id
}

func (m moduleNode) Value() string {
	return filepath.Base(m.path)
}

func (m moduleNode) String() string {
	return tui.ModulePathWithIcon(m.Value(), false)
}

type workspaceNode struct {
	id            resource.MonotonicID
	name          string
	current       bool
	resourceCount string
	cost          string
}

func (w workspaceNode) ID() any {
	return w.id
}

func (w workspaceNode) Value() string {
	return w.name
}

func (w workspaceNode) String() string {
	name := lipgloss.NewStyle().
		Render(w.name)
	s := tui.WorkspaceNameWithIcon(name, false)
	if w.current {
		s += lipgloss.NewStyle().
			Foreground(tui.LighterGrey).
			Render(" ✓")
	}
	if w.resourceCount != "" {
		s += lipgloss.NewStyle().
			Foreground(tui.LighterGrey).
			Italic(true).
			Render(fmt.Sprintf(" %s", w.resourceCount))
	}
	if w.cost != "" {
		s += lipgloss.NewStyle().
			Foreground(tui.Green).
			Italic(true).
			Render(fmt.Sprintf(" %s", w.cost))
	}
	return s
}
