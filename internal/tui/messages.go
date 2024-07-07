package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// NavigationMsg is an instruction to navigate to a page.
type NavigationMsg struct {
	Page Page
	Tag  int
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
		msg.Page.ID = parent.GetID()
	}
}

type InfoMsg string

// FilterFocusReqMsg is a request to focus the filter widget.
type FilterFocusReqMsg struct{}

// FilterBlurMsg is a request to unfocus the filter widget. It is not
// acknowledged.
type FilterBlurMsg struct{}

// FilterCloseMsg is a request to close the filter widget. It is not
// acknowledged.
type FilterCloseMsg struct{}

// FilterKeyMsg is a key entered by the user into the filter widget
type FilterKeyMsg tea.KeyMsg
