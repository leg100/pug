package task

import (
	"bufio"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
)

type machine struct {
	id       uuid.UUID
	viewport tui.Viewport
	spinner  *spinner.Model
	output   <-chan []byte
	buf      []byte
	scanner  *bufio.Scanner
}

type machineOptions struct {
	spinner *spinner.Model
	width   int
	height  int
}

func newMachine(t *task.Task, opts machineOptions) machine {
	return machine{
		id:      uuid.New(),
		spinner: opts.spinner,
		scanner: bufio.NewScanner(t.NewReader(false)),
		viewport: tui.NewViewport(tui.ViewportOptions{
			JSON:    t.JSON,
			Width:   opts.width,
			Height:  opts.height,
			Spinner: opts.spinner,
		}),
		output: t.NewStreamer(),
		// read upto 1kb at a time
		buf: make([]byte, 1024),
	}
}

func (h machine) Init() tea.Cmd {
	return h.getOutput
}

func (h machine) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.viewport.SetDimensions(msg.Width, msg.Height)
	case outputMsg:
		// Ensure output is for this model
		if msg.modelID != h.id {
			return h, nil
		}
		err := h.viewport.AppendContent(msg.output, msg.eof, true)
		if err != nil {
			return h, tui.ReportError(err)
		}
		if !msg.eof {
			return h, h.getOutput
		}
	default:
		// Handle keyboard and mouse events in the viewport
		var cmd tea.Cmd
		h.viewport, cmd = h.viewport.Update(msg)
		return h, cmd
	}
	return h, nil
}

func (h machine) View() string {
	return h.viewport.View()
}

func (h machine) getOutput() tea.Msg {
	if h.scanner.Scan() {
		h.scanner.Text
		msg := outputMsg{modelID: h.id}

	b, ok := <-h.output
	if ok {
		msg.output = b
	} else {
		msg.eof = true
	}
	return msg
}

type machineOutput struct{

}
