package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
)

const tabHeaderHeight = 2

// models implementing tabStatus can report a status that'll be rendered
// alongside the title in the tab header.
type tabStatus interface {
	TabStatus() string
}

// models implementing this can report info that'll be rendered on the opposite
// side from from the tab headers.
type tabSetInfo interface {
	// TabSetInfo is called with the name of the currently-active tab, or an
	// empty string if there are no tabs. It is expected to return a widget to
	// render.
	TabSetInfo(active tui.TabTitle) string
}

// A tab is one of a set of tabs. A tab has a title, and an embedded model,
// which is responsible for the visible content under the tab.
type Tab struct {
	tea.Model

	Title tui.TabTitle
}
