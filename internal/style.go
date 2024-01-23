package internal

import "github.com/charmbracelet/lipgloss"

var (
	regular        = lipgloss.NewStyle()
	roundedBorders = regular.Copy().Border(lipgloss.RoundedBorder())
)
