package tui

import "github.com/charmbracelet/lipgloss"

const (
	Black      = lipgloss.Color("#000000")
	Blue       = lipgloss.Color("63")
	DeepBlue   = lipgloss.Color("39")
	Pink       = lipgloss.Color("#E760FC")
	Darkred    = lipgloss.Color("#FF0000")
	LightGreen = lipgloss.Color("86")
	GreenBlue  = lipgloss.Color("#00A095")
	Green      = lipgloss.Color("34")
	DarkGreen  = lipgloss.Color("#325451")
	Grey       = lipgloss.Color("#737373")
	Orange     = lipgloss.Color("214")
	Red        = lipgloss.Color("204")
	Yellow     = lipgloss.Color("192")
	DarkGrey   = lipgloss.Color("#606362")
	White      = lipgloss.Color("#ffffff")
	OffWhite   = lipgloss.Color("#a8a7a5")
	LightGrey  = lipgloss.Color("245")
)

const (
	DebugLogLevel = Blue
	InfoLogLevel  = LightGreen
	ErrorLogLevel = Red
	WarnLogLevel  = Yellow
)
