package logs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/tui"
)

func coloredLogLevel(level string) string {
	var levelColor lipgloss.TerminalColor
	switch level {
	case "ERROR":
		levelColor = tui.ErrorLogLevel
	case "WARN":
		levelColor = tui.WarnLogLevel
	case "DEBUG":
		levelColor = tui.DebugLogLevel
	case "INFO":
		levelColor = tui.InfoLogLevel
	}
	return tui.Bold.Copy().Foreground(levelColor).Render(level)
}
