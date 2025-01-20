package task

import (
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
)

type human struct {
	id                uuid.UUID
	viewport          tui.Viewport
	spinner           *spinner.Model
	disableAutoscroll bool
	output            io.Reader
}

type humanOptions struct {
	disableAutoscroll bool
	spinner           *spinner.Model
	width             int
	height            int
}

func newHuman(t *task.Task, opts humanOptions) human {
	return human{
		id:      uuid.New(),
		spinner: opts.spinner,
		viewport: tui.NewViewport(tui.ViewportOptions{
			JSON:    t.JSON,
			Width:   opts.width,
			Height:  opts.height,
			Spinner: opts.spinner,
		}),
		// Disable autoscroll if either task is finished or user has disabled it
		disableAutoscroll: t.State.IsFinal() || opts.disableAutoscroll,
		output:            t.NewReader(true),
	}
}

func (h human) Init() tea.Cmd {
	return h.getOutput
}

func (h human) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.viewport.SetDimensions(msg.Width, msg.Height)
	case outputMsg:
		// Ensure output is for this model
		if msg.modelID != h.id {
			return h, nil
		}
		err := h.viewport.AppendContent(msg.output, msg.eof, !h.disableAutoscroll)
		if err != nil {
			return h, tui.ReportError(err)
		}
		if !msg.eof {
			return h, h.getOutput
		}
	case toggleAutoscrollMsg:
		h.disableAutoscroll = !h.disableAutoscroll
	default:
		// Handle keyboard and mouse events in the viewport
		var cmd tea.Cmd
		h.viewport, cmd = h.viewport.Update(msg)
		return h, cmd
	}
	return h, nil
}

func (h human) View() string {
	return h.viewport.View()
}

func (h human) getOutput() tea.Msg {
	msg := outputMsg{
		modelID: h.id,
		output:  make([]byte, 1024),
	}

	_, err := h.output.Read(msg.output)
	if errors.Is(err, io.EOF) {
		msg.eof = true
	} else if err != nil {
		return tui.ErrorMsg(fmt.Errorf("reading task output: %w", err))
	}

	return msg
}

type outputMsg struct {
	modelID uuid.UUID
	output  []byte
	eof     bool
}
