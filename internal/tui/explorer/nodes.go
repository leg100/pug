package explorer

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

const (
	dirIcon       string = ""
	moduleIcon    string = "󰠱"
	workspaceIcon string = ""
)

type node interface {
	fmt.Stringer

	// ID uniquely identifies the node
	ID() any
}

type nodeID any

type dirNode struct {
	path   string
	root   bool
	closed bool
}

func (d dirNode) ID() any {
	return nodeID(d.path)
}

func (d dirNode) String() string {
	if d.root {
		return fmt.Sprintf("%s %s", dirIcon, d.path)
	} else {
		return fmt.Sprintf("%s %s", dirIcon, filepath.Base(d.path))
	}
}

type moduleNode struct {
	id   resource.ID
	path string
}

func (m moduleNode) ID() any {
	return nodeID(m.id)
}

func (m moduleNode) String() string {
	return fmt.Sprintf("%s %s", moduleIcon, filepath.Base(m.path))
}

type workspaceNode struct {
	id            resource.ID
	name          string
	current       bool
	resourceCount string
}

func (w workspaceNode) ID() any {
	return nodeID(w.id)
}

func (w workspaceNode) String() string {
	name := lipgloss.NewStyle().
		Bold(w.current).
		Render(w.name)
	s := fmt.Sprintf("%s %s", workspaceIcon, name)
	if w.resourceCount != "" {
		s += lipgloss.NewStyle().
			Foreground(tui.LighterGrey).
			Italic(true).
			Render(fmt.Sprintf(" %s", w.resourceCount))
	}
	return s
}
