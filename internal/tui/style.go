package tui

import "github.com/charmbracelet/lipgloss"

var (
	Regular = lipgloss.NewStyle()
	Bold    = Regular.Copy().Bold(true)
	Padded  = Regular.Copy().Padding(0, 1)
	Faint   = Regular.Copy().Faint(true)

	Width  = lipgloss.Width
	Height = lipgloss.Height

	Border      = Regular.Copy().Border(lipgloss.NormalBorder())
	ThickBorder = Regular.Copy().Border(lipgloss.ThickBorder()).BorderForeground(Violet)

	ActiveBorder   = ThickBorder.Copy()
	InactiveBorder = Border.Copy().BorderForeground(LighterGrey)
)
