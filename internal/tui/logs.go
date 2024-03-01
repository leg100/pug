package tui

import (
	"bufio"
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	iopkg "github.com/leg100/pug/internal/io"
	"github.com/lmittmann/tint"
	"github.com/muesli/reflow/wordwrap"
)

const logsState = "logs"

type logMsg string

type model struct {
	current, last Page

	buf *iopkg.Buffer

	content  string
	viewport viewport.Model
}

func newLogs() model {
	buf := iopkg.NewBuffer()
	logger := slog.New(tint.NewHandler(buf, &tint.Options{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	slog.Warn("enabled tinted logging")

	return model{
		buf:      buf,
		viewport: viewport.New(0, 0),
	}
}

func (m model) Init() tea.Cmd {
	return m.readLogs
}

func (m model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case globalKeyMsg:
		switch {
		case key.Matches(msg.KeyMsg, Keys.Logs):
			if msg.Current != logsState {
				// open logs, keeping reference to last state
				m.last = m.current
				return m, navigate(logsState)
			}
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Escape):
			// close logs and return to last state
			return m, navigate(m.last)
		}
	case logMsg:
		// because log messages are line terminated, each one can be safely
		// word-wrapped and prepended to previously word-wrapped content.
		m.content = wordwrap.String(string(msg), m.viewport.Width) + m.content
		m.viewport.SetContent(m.content)
		m.viewport.GotoTop()
		return m, m.readLogs
	case errorMsg:
		args := append(msg.Args, "error", msg.Error.Error())
		slog.Error(msg.Message, args...)
	case viewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height
		// re-wrap entire content
		m.viewport.SetContent(wordwrap.String(m.content, m.viewport.Width))
		// Is this necessary?
		//m.viewport.GotoTop()
	case navigationMsg:
		m.current = msg.To
	default:
		// Handle keyboard and mouse events in the viewport
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return "logs"
}

func (m model) View() string {
	return lipgloss.NewStyle().
		Margin(0, 1).
		Render(m.viewport.View())
}

func (m model) readLogs() tea.Msg {
	b := new(strings.Builder)
	scanner := bufio.NewScanner(m.buf)
	for scanner.Scan() {
		fmt.Fprintln(b, scanner.Text()) // Println will add back the final '\n'
	}
	if err := scanner.Err(); err != nil {
		return logMsg(err.Error())
	}
	return logMsg(b.String())
}
