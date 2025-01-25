package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/hokaccha/go-prettyjson"
	"github.com/leg100/pug/internal/tui/keys"
)

// Viewport is a wrapper of the upstream viewport bubble.
type Viewport struct {
	viewport viewport.Model

	content           []byte
	json              bool
	multijson         bool
	spinner           *spinner.Model
	disableAutoscroll bool
}

type ViewportOptions struct {
	Width  int
	Height int
	// JSON is true if the content is a json object
	JSON bool
	// MultiJSON is true if the content is composed of multiple json objects,
	// one per line.
	MultiJSON         bool
	Border            bool
	Autoscroll        bool
	Spinner           *spinner.Model
	DisableAutoscroll bool
}

func NewViewport(opts ViewportOptions) Viewport {
	m := Viewport{
		viewport:          viewport.New(0, 0),
		json:              opts.JSON,
		multijson:         opts.MultiJSON,
		spinner:           opts.Spinner,
		disableAutoscroll: opts.DisableAutoscroll,
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
	case ToggleAutoscrollMsg:
		m.disableAutoscroll = !m.disableAutoscroll
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
	if m.multijson {
		if prettified, fmterr := prettyjson.Format(content); fmterr != nil {
			// In the event of an error, still set unprettified content
			// below.
			err = fmt.Errorf("pretty printing json content: %w", fmterr)
		} else {
			content = append(prettified, []byte("\n")...)
		}
	}
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
	if !m.disableAutoscroll {
		m.viewport.GotoBottom()
	}
	return err
}

func (m *Viewport) setContent() {
	// Wrap content to the width of the viewport, whilst respecting ANSI escape
	// codes (i.e. don't split codes across lines).
	wrapped := ansi.Wrap(ansi.Wordwrap(string(m.content), m.viewport.Width, ""), m.viewport.Width, "")
	sanitized := SanitizeColors([]byte(wrapped))
	m.viewport.SetContent(string(sanitized))
}

// ToggleAutoscrollMsg toggles whether viewport is auto-scrolled.
type ToggleAutoscrollMsg struct{}
