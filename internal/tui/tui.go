package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

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
	TaskKind
	LogsKind
)

// page identifies an instance of a model
type page struct {
	kind     modelKind
	resource resource.Resource
}

func (p page) cacheKey() cacheKey {
	return cacheKey{kind: p.kind, id: p.resource.ID()}
}
func cmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type errorMsg struct {
	Error   error
	Message string
	Args    []any
}

func newErrorMsg(err error, msg string, args ...any) errorMsg {
	return errorMsg{
		Error:   err,
		Message: msg,
		Args:    args,
	}
}

func newErrorCmd(err error, msg string, args ...any) tea.Cmd {
	return cmdHandler(newErrorMsg(err, msg, args...))
}

type ViewSizeMsg struct {
	Width, Height int
}
