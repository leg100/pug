package tui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/common"
	"github.com/muesli/reflow/wordwrap"
)

type logMsg string

type logsModel struct {
	lastPage     common.Page
	lastResource *resource.Resource

	content  string
	viewport viewport.Model
}

func newLogsModel() logsModel {
	return logsModel{
		viewport: viewport.New(0, 0),
	}
}

func (m logsModel) Init() tea.Cmd {
	return nil
}

func (m logsModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Escape):
			// return to last state
			return m, common.CmdHandler(common.ReturnLastMsg{})
		}
	case resource.Event[logging.Message]:
		// because log messages are line terminated, each one can be safely
		// word-wrapped and prepended to previously word-wrapped content.
		m.content = wordwrap.String(string(msg.Payload), m.viewport.Width) + m.content
		m.viewport.SetContent(m.content)
		m.viewport.GotoTop()
		return m, nil
	case errorMsg:
		args := append(msg.Args, "error", msg.Error.Error())
		slog.Error(msg.Message, args...)
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height
		// re-wrap entire content
		m.viewport.SetContent(wordwrap.String(m.content, m.viewport.Width))
		// Is this necessary?
		//m.viewport.GotoTop()
	}
	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m logsModel) Title() string {
	return "logs"
}

func (m logsModel) View() string {
	return lipgloss.NewStyle().
		Margin(0, 1).
		Render(m.viewport.View())
}

func (m logsModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}
