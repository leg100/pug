package navigator

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

// Go sends an instruction to navigate to a model.
func Go(kind tui.Kind, opts ...GoOption) tea.Cmd {
	return func() tea.Msg {
		msg := GoMsg(page{kind: kind, resource: resource.GlobalResource})
		for _, fn := range opts {
			fn(&msg)
		}
		return msg
	}
}

// GoMsg is an instruction to navigate to a page.
type GoMsg page

type GoOption func(msg *GoMsg)

func WithResource(res resource.Resource) GoOption {
	return func(msg *GoMsg) {
		msg.resource = res
	}
}

func WithTab(tab tui.TabTitle) GoOption {
	return func(msg *GoMsg) {
		msg.tab = tab
	}
}

// SwitchTabMsg is an instruction to switch tabs on the current page.
type SwitchTabMsg tui.TabTitle

func SwitchTab(tab tui.TabTitle) tea.Cmd {
	return func() tea.Msg {
		return SwitchTabMsg(tab)
	}
}
