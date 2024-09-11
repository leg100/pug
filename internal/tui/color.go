package tui

import "github.com/charmbracelet/lipgloss"

const (
	Black           = lipgloss.Color("#000000")
	DarkRed         = lipgloss.Color("#FF0000")
	Red             = lipgloss.Color("#FF5353")
	Purple          = lipgloss.Color("135")
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

	LogRecordAttributeKey = lipgloss.AdaptiveColor{Dark: string(LightGrey), Light: string(LightGrey)}

	HelpKey = lipgloss.AdaptiveColor{
		Dark:  "ff",
		Light: "",
	}
	HelpDesc = lipgloss.AdaptiveColor{
		Dark:  "248",
		Light: "246",
	}

	InactivePreviewBorder = lipgloss.AdaptiveColor{
		Dark:  "244",
		Light: "250",
	}

	CurrentBackground            = Grey
	CurrentForeground            = White
	SelectedBackground           = lipgloss.Color("110")
	SelectedForeground           = Black
	CurrentAndSelectedBackground = lipgloss.Color("117")
	CurrentAndSelectedForeground = Black

	TitleColor = lipgloss.AdaptiveColor{
		Dark:  "",
		Light: "",
	}

	modulePathColor = lipgloss.AdaptiveColor{
		Dark:  string(Grey),
		Light: string(Grey),
	}

	GroupReportBackgroundColor = EvenLighterGrey
	TaskSummaryBackgroundColor = EvenLighterGrey

	ScrollPercentageBackground = lipgloss.AdaptiveColor{
		Dark:  string(DarkGrey),
		Light: string(EvenLighterGrey),
	}
)
