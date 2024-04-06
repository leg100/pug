package tui

import "github.com/charmbracelet/lipgloss"

const (
	Black      = lipgloss.Color("#000000")
	DarkRed    = lipgloss.Color("#FF0000")
	Red        = lipgloss.Color("#FF5353")
	Pink       = lipgloss.Color("#E760FC")
	Orange     = lipgloss.Color("214")
	Yellow     = lipgloss.Color("#DBBD70")
	Green      = lipgloss.Color("34")
	LightGreen = lipgloss.Color("86")
	DarkGreen  = lipgloss.Color("#325451")
	GreenBlue  = lipgloss.Color("#00A095")
	DeepBlue   = lipgloss.Color("39")
	Blue       = lipgloss.Color("63")
	Grey       = lipgloss.Color("#737373")
	LightGrey  = lipgloss.Color("245")
	DarkGrey   = lipgloss.Color("#606362")
	White      = lipgloss.Color("#ffffff")
	OffWhite   = lipgloss.Color("#a8a7a5")
)

var (
	DebugLogLevel                    = Blue
	InfoLogLevel                     = lipgloss.AdaptiveColor{Dark: string(LightGreen), Light: string(Green)}
	ErrorLogLevel                    = Red
	WarnLogLevel                     = Yellow
	LogRecordAttributeKey            = lipgloss.AdaptiveColor{Dark: string(White), Light: string(Black)}
	HelpKey                          = Grey
	HelpDesc                         = LightGrey
	HighlightBackground              = Grey
	HighlightForeground              = White
	SelectedBackground               = lipgloss.Color("#a19518")
	SelectedForeground               = Black
	HighlightedAndSelectedBackground = lipgloss.Color("#635c0e")
	HighlightedAndSelectedForeground = White
)
