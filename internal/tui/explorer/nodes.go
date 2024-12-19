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
}

type dirNode struct {
	path   string
	root   bool
	closed bool
}

func (d dirNode) ID() any {
	return d.path
}

func (d dirNode) String() string {
	if d.root {
		return fmt.Sprintf("%s %s", tui.DirIcon, d.path)
	} else {
		return fmt.Sprintf("%s %s", tui.DirIcon, filepath.Base(d.path))
	}
}

type moduleNode struct {
	id   resource.ID
	path string
}

func (m moduleNode) ID() any {
	return m.id
}

func (m moduleNode) String() string {
	return tui.ModulePathWithIcon(filepath.Base(m.path), false)
}

type workspaceNode struct {
	id            resource.ID
	name          string
	current       bool
	resourceCount string
	cost          string
}

func (w workspaceNode) ID() any {
	return w.id
}

func (w workspaceNode) String() string {
	name := lipgloss.NewStyle().
		Bold(w.current).
		Render(w.name)
	s := tui.WorkspaceNameWithIcon(name, false)
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
			Render(fmt.Sprintf(" $%s", w.cost))
	}
	return s
}
