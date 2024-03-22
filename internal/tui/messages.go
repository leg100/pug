package tui

// NavigationMsg is an instruction to navigate to a page.
type NavigationMsg Page

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

// BodyResizeMsg is sent whenever the user resizes the terminal window. The width
// and height refer to area available in the main body between the header and
// the footer.
type BodyResizeMsg struct {
	Width, Height int
}
