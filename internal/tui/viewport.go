package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hokaccha/go-prettyjson"
	"github.com/leg100/reflow/wrap"
)

// Viewport is a wrapper of the upstream viewport bubble.
type Viewport struct {
	viewport viewport.Model

	Autoscroll bool

	border  bool
	content string
	json    bool
}

type ViewportOptions struct {
	Width      int
	Height     int
	JSON       bool
	Border     bool
	Autoscroll bool
}

func NewViewport(opts ViewportOptions) Viewport {
	m := Viewport{
		viewport:   viewport.New(0, 0),
		json:       opts.JSON,
		border:     opts.Border,
		Autoscroll: opts.Autoscroll,
	}
	m.SetDimensions(opts.Width, opts.Height)
	return m
}

func (m Viewport) Init() tea.Cmd {
	return nil
}

func (m Viewport) Update(msg tea.Msg) (Viewport, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Viewport) View() string {
	if m.border {
		return Border.Render(m.viewport.View())
	}
	return m.viewport.View()
}

func (m Viewport) SetDimensions(width, height int) {
	if m.border {
		// Subtract 2 to accomodate border
		width -= 2
		height -= 2
	}

	// If width has changed, re-wrap existing content.
	if m.viewport.Width != width {
		m.setContent()
	}

	m.viewport.Width = width
	m.viewport.Height = height
}

func (m *Viewport) AppendContent(content string, finished bool) error {
	m.content += content
	if finished {
		if m.content == "" {
			m.content = "No output"
		} else if m.json {
			// Prettify JSON output from task. This can only be done once
			// the task has finished and has produced complete and
			// syntactically valid json object(s).
			//
			// TODO: avoid casting to string and back, thereby avoiding
			// unnecessary allocations.
			if b, err := prettyjson.Format([]byte(m.content)); err != nil {
				return fmt.Errorf("pretty printing json content: %w", err)
			} else {
				m.content = string(b)
			}
		}
	}
	m.setContent()
	if m.Autoscroll {
		m.viewport.GotoBottom()
	}
	return nil
}

func (m *Viewport) setContent() {
	// TODO: set min width
	wrapped := wrap.String(m.content, m.viewport.Width)
	m.viewport.SetContent(wrapped)
}

func (m *Viewport) ToggleAutoscroll() {
	// TODO: set min width
	wrapped := wrap.String(m.content, m.viewport.Width)
	m.viewport.SetContent(wrapped)
}
