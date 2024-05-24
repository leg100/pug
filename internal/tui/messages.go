package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

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
