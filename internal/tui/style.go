package tui

import "github.com/charmbracelet/lipgloss"

var (
	Regular = lipgloss.NewStyle()
	Bold    = Regular.Bold(true)
	Padded  = Regular.Padding(0, 1)

	Border      = Regular.Border(lipgloss.NormalBorder())
	ThickBorder = Regular.Border(lipgloss.ThickBorder()).BorderForeground(Violet)

	ModuleStyle    = Regular.Foreground(DarkishGreen)
	WorkspaceStyle = Regular.Foreground(Purple)
)
