package tui

import "github.com/charmbracelet/lipgloss"

var (
	Regular = lipgloss.NewStyle()
	Bold    = Regular.Bold(true)
	Padded  = Regular.Padding(0, 1)

	Width  = lipgloss.Width
	Height = lipgloss.Height

	Border      = Regular.Border(lipgloss.NormalBorder())
	ThickBorder = Regular.Border(lipgloss.ThickBorder()).BorderForeground(Violet)

	Title          = Padded.Foreground(White).Background(Purple)
	TitleCommand   = Padded.Foreground(White).Background(Blue)
	TitlePath      = Padded.Foreground(White).Background(modulePathColor)
	TitleWorkspace = Padded.Foreground(White).Background(Green)
	TitleID        = Padded.Foreground(White).Background(Green)
	TitleAddress   = Padded.Foreground(White).Background(Blue)
	TitleSerial    = Padded.Foreground(Black).Background(Orange)
	TitleTainted   = Padded.Foreground(White).Background(Red)
)
