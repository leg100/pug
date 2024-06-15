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

	Title          = Regular.Copy().Padding(0, 1, 0, 0).Background(Pink).Foreground(White)
	TitleCommand   = Regular.Copy().Padding(0, 1).Foreground(White).Background(Blue)
	TitlePath      = Regular.Copy().Padding(0, 1).Foreground(White).Background(modulePathColor)
	TitleWorkspace = Regular.Copy().Padding(0, 1).Foreground(White).Background(Red)
	TitleID        = Regular.Copy().Padding(0, 1).Foreground(White).Background(Green)
	TitleAddress   = Regular.Copy().Padding(0, 1).Foreground(White).Background(Blue)
	TitleSerial    = Regular.Copy().Padding(0, 1).Foreground(White).Background(Orange)
)
