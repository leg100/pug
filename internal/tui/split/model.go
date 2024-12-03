package split

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

const (
	// default height of the top pane when split, including borders.
	defaultTopPaneHeight = 15
	// minimum height of the top pane when split, including borders.
	minTopPaneHeight = 10
	// minimum height of the bottom pane, including borders.
	minBottomPaneHeight = tui.MinContentHeight - minTopPaneHeight
)

type previewState int

const (
	previewHidden previewState = iota
	previewUnfocused
	previewFocused
)

// Model is a composition of two models corresponding to two panes: a top pane
// is a list of resources; the bottom pane provides further details of the
// resource corresponding to the current row in the list - this pane is known as
// the 'preview'.
type Model[R resource.Resource] struct {
	Table table.Model[R]
	maker tui.Maker

	previewState previewState
	height       int
	width        int
	focused      bool
	// topSplitHeight is the height of the top pane when the terminal is split.
	topSplitHeight int
	// cache of models for the previews
	cache *cache
}

type Options[R resource.Resource] struct {
	Columns      []table.Column
	Renderer     table.RowRenderer[R]
	TableOptions []table.Option[R]
	Width        int
	Height       int
	Maker        tui.Maker
}

func New[R resource.Resource](opts Options[R]) Model[R] {
	m := Model[R]{
		width:          opts.Width,
		height:         opts.Height,
		maker:          opts.Maker,
		cache:          newCache(),
		previewState:   previewUnfocused,
		topSplitHeight: defaultTopPaneHeight,
	}
	m.updateTopSplitHeight(0)
	// Create table for the top list pane
	m.Table = table.New(
		opts.Columns,
		opts.Renderer,
		opts.Width,
		m.tableHeight(),
		opts.TableOptions...,
	)
	return m
}

func (m Model[R]) Init() tea.Cmd {
	return nil
}

func (m Model[R]) Update(msg tea.Msg) (Model[R], tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Key handlers regardless of which pane is focused
		switch {
		case key.Matches(msg, Keys.SwitchPane):
			switch m.previewState {
			case previewUnfocused:
				m.previewState = previewFocused
			case previewFocused:
				m.previewState = previewUnfocused
			case previewHidden:
				return m, tui.CmdHandler(tui.FocusExplorerMsg{})
			}
			// m.setBorderStyles()
		case key.Matches(msg, Keys.ToggleSplit):
			switch m.previewState {
			case previewUnfocused, previewFocused:
				m.previewState = previewHidden
			case previewHidden:
				m.previewState = previewUnfocused
			}
			m.propagateDimensions()
		case key.Matches(msg, Keys.IncreaseSplit):
			m.updateTopSplitHeight(1)
		case key.Matches(msg, Keys.DecreaseSplit):
			m.updateTopSplitHeight(-1)
		}
		if m.previewState == previewFocused {
			// Preview pane is focused, so send keys to the preview
			// model for the currently highlighted table row if there is one.
			row, ok := m.Table.CurrentRow()
			if !ok {
				break
			}
			cmds = append(cmds, m.cache.Update(row.ID, msg))
		} else {
			// Table pane is focused, so handle keys relevant to table rows.
			m.Table, cmd = m.Table.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.propagateDimensions()
	default:
		// Forward remaining message types to both the table model and cached
		// resource models
		m.Table, cmd = m.Table.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.cache.UpdateAll(msg)...)
	}
	if m.previewState != previewHidden {
		// Get current table row and ensure a model exists for it, and
		// ensure that that model is the current model.
		if row, ok := m.Table.CurrentRow(); ok {
			if model := m.cache.Get(row.ID); model == nil {
				// Create model
				model, err := m.maker.Make(row.ID, m.contentWidth(), m.previewHeight())
				if err != nil {
					return m, tui.ReportError(fmt.Errorf("making model for preview: %w", err))
				}
				// Cache newly created model
				m.cache.Put(row.ID, model)
				// Initialize model
				cmds = append(cmds, model.Init())
			}
		}
	}
	return m, tea.Batch(cmds...)
}

// propagateDimensions propagates respective dimensions to child models
func (m *Model[R]) propagateDimensions() {
	m.Table, _ = m.Table.Update(tea.WindowSizeMsg{
		Height: m.tableHeight(),
		Width:  m.contentWidth(),
	})
	_ = m.cache.UpdateAll(tea.WindowSizeMsg{
		Height: m.previewHeight(),
		Width:  m.contentWidth(),
	})
}

func (m *Model[R]) updateTopSplitHeight(delta int) {
	switch m.previewState {
	case previewHidden:
		return
	case previewFocused:
		delta = -delta
	}
	m.topSplitHeight = clamp(m.topSplitHeight+delta, minTopPaneHeight, m.height-minBottomPaneHeight)
	m.propagateDimensions()
}

// tableHeight returns the height of the table within the top pane border
func (m Model[R]) tableHeight() int {
	var h int
	if m.previewState != previewHidden {
		h = m.topSplitHeight
	} else {
		h = m.height
	}
	// minus 2 to accomodate borders
	return max(0, h-2)
}

// previewHeight returns the height of the preview within the bottom pane border
func (m Model[R]) previewHeight() int {
	h := max(minBottomPaneHeight, m.height-m.topSplitHeight)
	// minus 2 to accomodate borders
	return max(0, h-2)
}

// contentWidth returns the width of the content within the borders of a pane.
func (m *Model[R]) contentWidth() int {
	// minus 2 to accomodate borders
	return m.width - 2
}

func (m Model[R]) View() string {
	var (
		model     tui.ChildModel
		onlyTable bool
	)
	// Render only the table if the preview is hidden, or there is no model
	// corresponding to the preview pane.
	if m.previewState == previewHidden {
		onlyTable = true
	} else {
		var ok bool
		model, ok = m.getPreviewModel()
		onlyTable = !ok
	}
	if onlyTable {
		return lipgloss.NewStyle().
			Border(tui.BorderStyle(m.focused)).
			BorderForeground(tui.BorderColor(m.focused)).
			Render(m.Table.View())
	}
	tbl := lipgloss.NewStyle().
		Border(tui.BorderStyle(m.previewState == previewUnfocused)).
		BorderForeground(tui.BorderColor(m.focused && m.previewState == previewUnfocused)).
		Render(m.Table.View())
	preview := lipgloss.NewStyle().
		Border(tui.BorderStyle(m.previewState == previewFocused)).
		BorderForeground(tui.BorderColor(m.focused && m.previewState == previewFocused)).
		Render(model.View())
	return lipgloss.JoinVertical(lipgloss.Top, tbl, preview)
}

func (m *Model[R]) Focus(focused bool) {
	m.focused = focused
}

// getPreviewModel returns the model for the preview pane.
func (m Model[R]) getPreviewModel() (tui.ChildModel, bool) {
	row, ok := m.Table.CurrentRow()
	if !ok {
		return nil, false
	}
	model := m.cache.Get(row.ID)
	return model, model != nil
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
