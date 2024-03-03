package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// TODO: rename to ModelKind?

type Page int

const (
	ModuleListPage Page = iota
	WorkspaceListPage
	RunListPage
	TaskListPage
	TaskPage
	LogsPage
	HelpPage
)

type NavigationMsg struct {
	To       Page
	Resource *resource.Resource
}

func Navigate(to Page, res *resource.Resource) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{To: to, Resource: res}
	}
}

// ReturnLastMsg is sent when the user wants to return to the last page.
type ReturnLastMsg struct{}
