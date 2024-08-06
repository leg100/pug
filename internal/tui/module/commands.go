package module

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/tui"
)

// ReloadModules reloads pug modules, resolving any differences between the
// modules on the user's disk, and those loaded in pug. Set firsttime to toggle
// whether this is the first time modules are being loaded.
func ReloadModules(firsttime bool, modules tui.ModuleService) tea.Cmd {
	return func() tea.Msg {
		// TODO: pass in bubbletea context once supported.
		loadTotal, unloadTotal := calculateReloadStats(modules.Reload(context.TODO()))
		if firsttime {
			return tui.InfoMsg(
				fmt.Sprintf("loaded %d modules", loadTotal),
			)
		} else {
			return tui.InfoMsg(
				fmt.Sprintf("reloaded modules: added %d; removed %d", loadTotal, unloadTotal),
			)
		}
	}
}

func calculateReloadStats(results chan module.ReloadResult) (loadTotal int, unloadTotal int) {
	for result := range results {
		if result.Loaded {
			loadTotal++
		} else {
			unloadTotal++
		}
	}
	return
}
