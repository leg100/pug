package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/reflow/wordwrap"
	"github.com/leg100/reflow/wrap"
)

// Viewport is a wrapper of the upstream viewport bubble.
type Viewport struct {
	viewport viewport.Model

	Autoscroll bool

	content []byte
	json    bool
	spinner *spinner.Model
}

type ViewportOptions struct {
	Width      int
	Height     int
	JSON       bool
	Border     bool
	Autoscroll bool
	Spinner    *spinner.Model
}

func NewViewport(opts ViewportOptions) Viewport {
	m := Viewport{
		Autoscroll: opts.Autoscroll,
		viewport:   viewport.New(0, 0),
		json:       opts.JSON,
		spinner:    opts.Spinner,
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Navigation.GotoTop):
			m.viewport.SetYOffset(0)
		case key.Matches(msg, keys.Navigation.GotoBottom):
			m.viewport.SetYOffset(m.viewport.TotalLineCount())
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Viewport) View() string {
	var output string
	if len(m.content) == 0 {
		msg := "Awaiting output"
		// TODO: make spinner non-optional
		if m.spinner != nil {
			msg += " " + m.spinner.View()
		}
		output = Regular.
			Height(m.viewport.Height).
			Width(m.viewport.Width).
			Render(msg)
	} else {
		output = m.viewport.View()
	}
	scrollbar := Scrollbar(
		m.viewport.Height,
		m.viewport.TotalLineCount(),
		m.viewport.VisibleLineCount(),
		m.viewport.YOffset,
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, output, scrollbar)
}

func (m *Viewport) SetDimensions(width, height int) {
	width = max(0, width-ScrollbarWidth)
	// If width has changed, re-wrap existing content.
	rewrap := m.viewport.Width != width
	m.viewport.Width = width
	m.viewport.Height = height
	if rewrap {
		m.setContent()
	}
}

func (m *Viewport) AppendContent(content []byte, finished bool) (err error) {
	m.content = append(m.content, content...)
	if finished {
		if len(m.content) == 0 {
			m.content = []byte("No output")
		} else if m.json {
			// Prettify JSON output from task. This can only be done once
			// the task has finished and has produced complete and
			// syntactically valid json object(s).
			if b, fmterr := prettyjson.Format(m.content); fmterr != nil {
				// In the event of an error, still set unprettified content
				// below.
				err = fmt.Errorf("pretty printing json content: %w", fmterr)
			} else {
				m.content = b
			}
		}
	}
	m.setContent()
	if m.Autoscroll {
		m.viewport.GotoBottom()
	}
	return err
}

func (m *Viewport) setContent() {
	// Wrap content to the width of the viewport, whilst respecting ANSI escape
	// codes (i.e. don't split codes across lines). The wrapper also ensures
	// thekkkk
	wrapped := wrap.Bytes(wordwrap.Bytes(m.content, m.viewport.Width), m.viewport.Width)
	sanitized := SanitizeColors(wrapped)
	m.viewport.SetContent(string(sanitized))
}
