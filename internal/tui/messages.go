package tui

import (
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
)

// NavigationMsg is an instruction to navigate to a page.
type NavigationMsg struct {
	Page     Page
	TabTitle string
}

func NewNavigationMsg(kind Kind, opts ...NavigateOption) NavigationMsg {
	msg := NavigationMsg{Page: Page{Kind: kind}}
	for _, fn := range opts {
		fn(&msg)
	}
	return msg
}

type NavigateOption func(msg *NavigationMsg)

func WithParent(parent resource.Resource) NavigateOption {
	return func(msg *NavigationMsg) {
		msg.Page.Parent = parent
	}
}

type SetActiveTabMsg string

type InfoMsg string

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

type CreatedTasksMsg struct {
	// The command of the completed tasks (all tasks are assumed to be running
	// the same command).
	Command string
	Tasks   task.Multi
	// Errors from creating tasks
	CreateErrs []error
}

type CompletedTasksMsg struct {
	// The command of the completed tasks (all tasks are assumed to be running
	// the same command).
	Command string
	Tasks   task.Multi
	// Errors from originally creating tasks
	CreateErrs []error
}

type CreatedRunsMsg struct {
	Runs []*run.Run
	// Errors from creating tasks
	CreateErrs []error
}

// BodyResizeMsg is sent whenever the user resizes the terminal window. The width
// and height refer to area available in the main body between the header and
// the footer.
type BodyResizeMsg struct {
	Width, Height int
}
