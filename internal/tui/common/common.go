package common

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (Model, tea.Cmd)
	// Title is the single line bar that separates the header from the content.
	//
	// TODO: rename to breadcrumbs
	Title() string
	View() string
	// HelpBindings are those bindings that help should show when this model is
	// current.
	HelpBindings() []key.Binding
}

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type ErrorMsg struct {
	Error   error
	Message string
	Args    []any
}

func NewErrorMsg(err error, msg string, args ...any) ErrorMsg {
	return ErrorMsg{
		Error:   err,
		Message: msg,
		Args:    args,
	}
}

func NewErrorCmd(err error, msg string, args ...any) tea.Cmd {
	return CmdHandler(NewErrorMsg(err, msg, args...))
}

type ViewSizeMsg struct {
	Width, Height int
}

type BulkInsertMsg[T any] []T
