package tui

import "github.com/charmbracelet/lipgloss"

var (
	Regular        = lipgloss.NewStyle()
	RoundedBorders = Regular.Copy().Border(lipgloss.RoundedBorder())
	Bold           = Regular.Copy().Bold(true)

	width  = lipgloss.Width
	height = lipgloss.Height
)
