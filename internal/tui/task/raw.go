package task

import (
	"bufio"
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
)

type raw struct {
	id                uuid.UUID
	viewport          tui.Viewport
	scanner           *bufio.Scanner
	spinner           *spinner.Model
	disableAutoscroll bool
	output            io.Reader
}

type rawOptions struct {
	disableAutoscroll bool
	spinner           *spinner.Model
	width             int
	height            int
}

func newRaw(t *task.Task, opts rawOptions) raw {
	return raw{
		id:      uuid.New(),
		spinner: opts.spinner,
		viewport: tui.NewViewport(tui.ViewportOptions{
			MultiJSON: true,
			Width:     opts.width,
			Height:    opts.height,
			Spinner:   opts.spinner,
		}),
		scanner: bufio.NewScanner(t.NewReader(false)),
		// Disable autoscroll if either task is finished or user has disabled it
		disableAutoscroll: t.State.IsFinal() || opts.disableAutoscroll,
		output:            t.NewReader(true),
	}
}

func (m raw) Init() tea.Cmd {
	return m.getOutput
}

func (m raw) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.SetDimensions(msg.Width, msg.Height)
	case outputMsg:
		// Ensure output is for this model
		if msg.modelID != m.id {
			return m, nil
		}
		if msg.eof {
			return m, nil
		}
		err := m.viewport.AppendContent(msg.output, msg.eof)
		if err != nil {
			return m, tui.ReportError(err)
		}
		return m, m.getOutput
	default:
		// Handle keyboard and mouse events in the viewport
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m raw) View() string {
	return m.viewport.View()
}

func (m raw) getOutput() tea.Msg {
	msg := outputMsg{modelID: m.id}
	if m.scanner.Scan() {
		msg.output = m.scanner.Bytes()
	} else {
		msg.eof = true
	}
	return msg
}
