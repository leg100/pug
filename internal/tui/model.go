package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/maps"
)

// maker makes new models
type maker interface {
	makeModel(target resource.Resource) (Model, error)
}

// Model essentially wraps the upstream tea.Model with additional methods.
type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (Model, tea.Cmd)
	Title() string
	View() string
	// Pagination renders pagination/scrolling info in the bottom right corner.
	Pagination() string
	// HelpBindings are those bindings that help should show when this model is
	// current.
	HelpBindings() []key.Binding
}

type modelKind int

const (
	ModuleListKind modelKind = iota
	WorkspaceListKind
	RunListKind
	TaskListKind
	RunKind
	TaskKind
	LogsKind
)

var firstPages = map[string]modelKind{
	"modules":    ModuleListKind,
	"workspaces": WorkspaceListKind,
	"runs":       RunListKind,
	"tasks":      TaskListKind,
}

// firstPageKind retrieves the model corresponding to the user requested first
// page.
func firstPageKind(s string) (modelKind, error) {
	kind, ok := firstPages[s]
	if !ok {
		return 0, fmt.Errorf("invalid first page, must be one of: %v", maps.Keys(firstPages))
	}
	return kind, nil
}
