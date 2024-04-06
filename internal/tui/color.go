package tui

import "github.com/charmbracelet/lipgloss"

const (
	Black           = lipgloss.Color("#000000")
	DarkRed         = lipgloss.Color("#FF0000")
	Red             = lipgloss.Color("#FF5353")
	Pink            = lipgloss.Color("#E760FC")
	Orange          = lipgloss.Color("214")
	Yellow          = lipgloss.Color("#DBBD70")
	Green           = lipgloss.Color("34")
	LightGreen      = lipgloss.Color("86")
	DarkGreen       = lipgloss.Color("#325451")
	GreenBlue       = lipgloss.Color("#00A095")
	DeepBlue        = lipgloss.Color("39")
	LightBlue       = lipgloss.Color("81")
	Blue            = lipgloss.Color("63")
	Violet          = lipgloss.Color("13")
	Grey            = lipgloss.Color("#737373")
	LightGrey       = lipgloss.Color("245")
	LighterGrey     = lipgloss.Color("250")
	EvenLighterGrey = lipgloss.Color("253")
	DarkGrey        = lipgloss.Color("#606362")
	White           = lipgloss.Color("#ffffff")
	OffWhite        = lipgloss.Color("#a8a7a5")
)

var (
	DebugLogLevel = Blue
	InfoLogLevel  = lipgloss.AdaptiveColor{Dark: string(LightGreen), Light: string(Green)}
	ErrorLogLevel = Red
	WarnLogLevel  = Yellow

	LogRecordAttributeKey = lipgloss.AdaptiveColor{Dark: string(White), Light: string(Black)}

	HelpKey  = Grey
	HelpDesc = LightGrey

	HighlightBackground              = Grey
	HighlightForeground              = White
	SelectedBackground               = lipgloss.Color("#a19518")
	SelectedForeground               = Black
	HighlightedAndSelectedBackground = lipgloss.Color("#635c0e")
	HighlightedAndSelectedForeground = White

	TitleColor = lipgloss.AdaptiveColor{
		Dark:  "",
		Light: "",
	}

	modulePathColor = lipgloss.AdaptiveColor{
		Dark:  string(Pink),
		Light: string(Pink),
	}

	globalColor = lipgloss.AdaptiveColor{
		Dark:  string(Pink),
		Light: string(Pink),
	}

	activeTabColor = lipgloss.AdaptiveColor{
		Dark:  string(LightBlue),
		Light: string(Violet),
	}
	inactiveTabColor = lipgloss.AdaptiveColor{
		Dark:  string(DarkGrey),
		Light: string(LighterGrey),
	}

	ScrollPercentageBackground = lipgloss.AdaptiveColor{
		Dark:  string(DarkGrey),
		Light: string(EvenLighterGrey),
	}
)
