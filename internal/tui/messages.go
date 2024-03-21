package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// NavigationMsg is an instruction to navigate to a page.
type NavigationMsg Page

// NavigateTo sends an instruction to navigate to a page with the given model
// kind, and optionally parent resource.
func NavigateTo(kind Kind, parent *resource.Resource) tea.Cmd {
	return func() tea.Msg {
		page := Page{Kind: kind}
		if parent != nil {
			page.Parent = *parent
		}
		return NavigationMsg(page)
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

// BodyResizeMsg is sent whenever the user resizes the terminal window. The width
// and height refer to area available in the main body between the header and
// the footer.
type BodyResizeMsg struct {
	Width, Height int
}
