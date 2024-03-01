package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type Page int

const (
	GlobalModuleListPage Page = iota
	GlobalWorkspaceListPage
	GlobalRunListPage
	GlobalTaskListPage
	ModuleWorkspaceListPage
	ModuleRunListPage
	ModuleTaskPage
	WorkspaceRunListPage
	RunTaskPage
)

type NavigationMsg struct {
	To Page
	ID uuid.UUID
}

func Navigate(to Page, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		return NavigationMsg{To: to, ID: id}
	}
}
