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

	Title          = Bold.Copy().Padding(0, 1).Background(Pink).Foreground(White)
	TitleCommand   = Padded.Copy().Foreground(White).Background(Blue)
	TitlePath      = Padded.Copy().Foreground(White).Background(modulePathColor)
	TitleWorkspace = Padded.Copy().Foreground(White).Background(Green)
	TitleID        = Padded.Copy().Foreground(White).Background(Green)
	TitleAddress   = Padded.Copy().Foreground(White).Background(Blue)
	TitleSerial    = Padded.Copy().Foreground(White).Background(Orange)
	TitleTainted   = Padded.Copy().Foreground(White).Background(Red)

	RunReportStyle = Padded.Copy().Background(EvenLighterGrey)
)
