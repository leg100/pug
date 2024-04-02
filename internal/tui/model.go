package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// Maker makes new models
type Maker interface {
	Make(target resource.Resource, width, height int) (Model, error)
}

// Model essentially wraps the upstream tea.Model with additional methods.
type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (Model, tea.Cmd)
	Title() string
	View() string
	// HelpBindings are those bindings that help should show when this model is
	// current.
	HelpBindings() []key.Binding
}

// Page identifies an instance of a model
type Page struct {
	// The model kind
	Kind Kind
	// The model's parent resource
	Parent resource.Resource
}

// ModelID is implemented by models that are able to provide a unique
// identification string.
type ModelID interface {
	ID() string
}

// ModelID is implemented by models that are able to provide a unique
// identification string.
type ModelStatus interface {
	Status() string
}
