package module

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
)

func ReloadModules(modules tui.ModuleService) tea.Cmd {
	return func() tea.Msg {
		if err := modules.Reload(); err != nil {
			return tui.NewErrorMsg(err, "reloading modules")
		}
		return nil
	}
}
