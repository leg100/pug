package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// Maker makes new models
type Maker interface {
	Make(target resource.Resource, width, height int) (tea.Model, error)
}

// Page identifies an instance of a model
type Page struct {
	// The model kind
	Kind Kind
	// The model's parent resource
	//
	// TODO: rename: it's sometimes a parent (in the case of listings), and
	// sometimes *the* resource.
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

// ModelTitle is implemented by models that show a title
type ModelTitle interface {
	Title() string
}

// ModelHelpBindings is implemented by models that surface further help bindings
// specific to the model.
type ModelHelpBindings interface {
	HelpBindings() []key.Binding
}
