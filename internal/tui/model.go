package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

type ChildModel interface {
	Init() tea.Cmd
	Update(tea.Msg) tea.Cmd
	View() string
}

// Maker makes new models
type Maker interface {
	Make(id resource.ID, width, height int) (ChildModel, error)
}

// Page identifies an instance of a model
type Page struct {
	// The model kind. Identifies the model maker to construct the page.
	Kind Kind
	// The ID of the resource for a model. In the case of global listings of
	// modules, workspaces, etc, this is the global resource.
	ID resource.ID
}

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
