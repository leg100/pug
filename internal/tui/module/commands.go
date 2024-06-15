package module

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
)

// ReloadModules reloads pug modules, resolving any differences between the
// modules on the user's disk, and those loaded in pug. Set firsttime to toggle
// whether this is the first time modules are being loaded.
func ReloadModules(firsttime bool, modules tui.ModuleService) tea.Cmd {
	return func() tea.Msg {
		added, removed, err := modules.Reload()
		if err != nil {
			return tui.ReportError(fmt.Errorf("reloading modules: %w", err))()
		}
		if firsttime {
			return tui.InfoMsg(
				fmt.Sprintf("loaded %d modules", len(added)),
			)
		} else {
			return tui.InfoMsg(
				fmt.Sprintf("reloaded modules: added %d; removed %d", len(added), len(removed)),
			)
		}
	}
}
