package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
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
