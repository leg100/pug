package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hokaccha/go-prettyjson"
	"github.com/leg100/reflow/wrap"
)

// viewportID is a global counter used to uniquely identify each viewport.
var viewportID int

// Viewport is a wrapper of the upstream viewport bubble enabling
// high-performance mode (HPM). HPM requires careful handling in non-trivial
// applications such as Pug, because it reserves a portion of the terminal for
// rendering to, which needs to be selectively turned off and on depending on
// whether the viewport is currently visible.
type Viewport struct {
	viewport.Model

	id      int
	json    bool
	sync    bool
	visible bool
}

type ViewportOptions struct {
	Width, Height int
	YPosition     int
	JSON          bool
}

// YPositionMsg is sent when the viewport's position on the Y-axis has changed,
// i.e. it has moved up or down with respect to the top of the terminal.
type YPositionMsg int

type clearViewportMsg struct{}

// ClearViewport clears the viewport scroll area. It should be sent whenever the
// currently visible model is changed, in case the previous model has a viewport
// that has reserved a portion of the terminal for rendering.
func ClearViewport() tea.Cmd {
	return tea.Batch(tea.ClearScrollArea, CmdHandler(clearViewportMsg{}))
}

// activateViewportMsg activates the viewport with the given ID. Once activated
// the viewport has exclusive rights to reserve a portion of the terminal for
// rendering.
type activateViewportMsg int

func NewViewport(opts ViewportOptions) Viewport {
	viewport := viewport.New(opts.Width, opts.Height)
	viewport.HighPerformanceRendering = true
	viewport.YPosition = opts.YPosition
	viewportID++
	return Viewport{
		Model: viewport,
		json:  opts.JSON,
		id:    viewportID,
	}
}

// Initialising the viewport first sends out a message to tell another viewport
// to stop reserving the terminal (if there is one). It then sends out messages
// to both synchronise and activate itself, to ensure any current content is
// rendered, and to ensure any future updates are rendered, respectively,
// whether they be window size changes, y position changes, or additional
// content.
func (m Viewport) Init() tea.Cmd {
	return tea.Sequence(CmdHandler(clearViewportMsg{}), tea.Batch(
		viewport.Sync(m.Model),
		CmdHandler(activateViewportMsg(m.id)),
	))
}

func (m Viewport) Update(msg tea.Msg) (Viewport, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.sync = true
	case YPositionMsg:
		m.YPosition = int(msg)
		m.sync = true
	case clearViewportMsg:
		m.visible = false
	case activateViewportMsg:
		if int(msg) == m.id {
			m.visible = true
		}
	}

	if m.sync && m.visible {
		cmds = append(cmds, viewport.Sync(m.Model))
		m.sync = false
	}

	// Handle keyboard and mouse events in the viewport
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Viewport) SetContent(content string, finished bool) error {
	if m.json && finished {
		// Prettify JSON content.
		//
		// TODO: avoid casting to string and back.
		if b, err := prettyjson.Format([]byte(content)); err != nil {
			return fmt.Errorf("pretty printing task json output: %w", err)
		} else {
			content = string(b)
		}
	}
	wrapped := wrap.String(content, m.Width)
	m.Model.SetContent(wrapped)
	// Whenever content is set on a viewport it should be re-synced. The caller
	// is responsible for ensuring the Update() method is called afterwards.
	m.sync = true
	return nil
}
