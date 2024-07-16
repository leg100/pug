package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/leg100/reflow/wrap"
)

// Viewport is a wrapper of the upstream viewport bubble.
type Viewport struct {
	viewport viewport.Model

	Autoscroll bool

	content string
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

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// percentWidth is the width of the scroll percentage section to the
// right of the viewport
const percentWidth = 6 // 6 = 4 for xxx% + 2 for padding

func (m Viewport) View() string {
	// scroll percent container occupies a fixed width section to the right of
	// the viewport.
	percent := Regular.
		Background(ScrollPercentageBackground).
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	percentContainer := Regular.
		Height(m.viewport.Height).
		Width(percentWidth).
		AlignVertical(lipgloss.Bottom).
		Render(percent)

	var output string
	if m.content == "" {
		msg := "Awaiting output"
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

	return lipgloss.JoinHorizontal(lipgloss.Top,
		output,
		percentContainer,
	)
}

func (m *Viewport) SetDimensions(width, height int) {
	width = max(0, width-percentWidth)
	// If width has changed, re-rewrap existing content.
	rewrap := m.viewport.Width != width
	m.viewport.Width = width
	m.viewport.Height = height
	if rewrap {
		m.setContent()
	}
}

func (m *Viewport) AppendContent(content string, finished bool) (err error) {
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
			if b, fmterr := prettyjson.Format([]byte(m.content)); fmterr != nil {
				// In the event of an error, still set unprettified content
				// below.
				err = fmt.Errorf("pretty printing json content: %w", fmterr)
			} else {
				m.content = string(b)
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
	wrapped := wrap.String(m.content, m.viewport.Width)
	m.viewport.SetContent(wrapped)
}
