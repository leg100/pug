package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// NavigationMsg is an instruction to navigate to a page.
type NavigationMsg struct {
	Page Page
	Tab  string
}

func NewNavigationMsg(kind Kind, opts ...NavigateOption) NavigationMsg {
	msg := NavigationMsg{Page: Page{Kind: kind, Resource: resource.GlobalResource}}
	for _, fn := range opts {
		fn(&msg)
	}
	return msg
}

type NavigateOption func(msg *NavigationMsg)

func WithParent(parent resource.Resource) NavigateOption {
	return func(msg *NavigationMsg) {
		msg.Page.Resource = parent
	}
}

func WithTab(tab string) NavigateOption {
	return func(msg *NavigationMsg) {
		msg.Tab = tab
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

// FilterFocusReqMsg is a request to focus the filter widget. FilterFocusAckMsg
// should be sent in response to ackowledge the request.
type FilterFocusReqMsg struct{}

// FilterFocusAckMsg acknowledges a request to focus the filter widget
type FilterFocusAckMsg struct{}

// FilterBlurMsg is a request to unfocus the filter widget. It is not
// acknowledged.
type FilterBlurMsg struct{}

// FilterCloseMsg is a request to close the filter widget. It is not
// acknowledged.
type FilterCloseMsg struct{}

// FilterKeyMsg is a key entered by the user into the filter widget
type FilterKeyMsg tea.KeyMsg
