package task

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
)

// Config is global task configuration
type Config struct {
	// disableAutoscroll disables auto-scrolling of task output.
	disableAutoscroll bool
	// showInfo shows further info about the task.
	showInfo bool
}

type (
	// toggleAutoscrollMsg toggles whether task output is auto-scrolled.
	toggleAutoscrollMsg struct{}

	// toggleShowInfo toggles whether task info is shown.
	toggleShowInfo struct{}
)

// Update updates global configuration of tasks.
func (c *Config) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Autoscroll):
			c.disableAutoscroll = !c.disableAutoscroll

			// Inform user, and send out message to all cached task models to
			// toggle autoscroll.
			return tea.Batch(
				tui.CmdHandler(toggleAutoscrollMsg{}),
				tui.ReportInfo(fmt.Sprintf("Toggled autoscroll %s", boolToOnOff(!c.disableAutoscroll))),
			)
		case key.Matches(msg, localKeys.ToggleInfo):
			c.showInfo = !c.showInfo

			// Send out message to all cached task models to toggle task info
			return tui.CmdHandler(toggleShowInfo{})
		}
	}
	return nil
}
