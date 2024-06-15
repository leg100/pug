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
	// The model kind. Identifies the model maker to construct the page.
	Kind Kind
	// The model's resource. In the case of global listings, this is the global
	// resource. For all other listings, this is the parent resource of the
	// listed resources.
	Resource resource.Resource
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
