package tui

import "github.com/charmbracelet/lipgloss"

var (
	Regular = lipgloss.NewStyle()
	Bold    = Regular.Copy().Bold(true)
	Padded  = Regular.Copy().Padding(0, 1)

	Width  = lipgloss.Width
	Height = lipgloss.Height
)
