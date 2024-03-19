package tui

import "github.com/charmbracelet/lipgloss"

var (
	Regular        = lipgloss.NewStyle()
	RoundedBorders = Regular.Copy().Border(lipgloss.RoundedBorder())
	Bold           = Regular.Copy().Bold(true)

	Width  = lipgloss.Width
	Height = lipgloss.Height
)
