package split

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
)

const (
	// default height of the top list pane, not including borders
	defaultListPaneHeight = 10
	// height of the divider line in between the two panes.
	dividerHeight = 1
)

type Options[R resource.Resource] struct {
	Columns      []table.Column
	Renderer     table.RowRenderer[R]
	TableOptions []table.Option[resource.ID, R]
	Width        int
	Height       int
	Maker        Maker
}

type Maker interface {
	MakePreview(res resource.Resource, width, height, yPos int) (tea.Model, error)
}

func New[R resource.Resource](opts Options[R]) Model[R] {
	m := Model[R]{
		width:          opts.Width,
		height:         opts.Height,
		maker:          opts.Maker,
		cache:          newCache(),
		previewVisible: true,
	}
	// Create table for the top list pane
	m.Table = table.New(
		opts.Columns,
		opts.Renderer,
		opts.Width,
		m.listHeight(),
		opts.TableOptions...,
	)
	return m
}

// Model is a composition of two models corresponding to two panes: a top pane
// is a list of resources; the bottom pane provides further details of the
// resource corresponding to the current row in the list - this pane is known as
// the 'preview'.
type Model[R resource.Resource] struct {
	Table table.Model[resource.ID, R]
	maker Maker

	previewVisible bool
	previewFocused bool
	height         int
	width          int

	// userListHeightAdjustment is the adjustment the user has requested to the
	// default height of the list pane.
	userListHeightAdjustment int

	// cache of models for the previews
	cache *cache

	currentRowID resource.ID
}

// Init initialises the split model, first sending out a message to tell any
// previously visible viewport to stop reserving the terminal. Secondly, if
// preview is enabled, and there is a preview model for the current table row,
// then it is initialised, to ensure that its viewport now reserves the
// terminal.
func (m Model[R]) Init() tea.Cmd {
	cmds := []tea.Cmd{tui.ClearViewport()}
	if m.previewVisible {
		if preview := m.getPreviewModel(); preview != nil {
			cmds = append(cmds, preview.Init())
		}
	}
	return tea.Sequence(cmds...)
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
		case key.Matches(msg, keys.Navigation.SwitchPane):
			m.previewFocused = !m.previewFocused
		case key.Matches(msg, localKeys.TogglePreview):
			m.previewVisible = !m.previewVisible
			m.recalculateDimensions()
			if m.previewVisible {
				// toggle visibility = true for current preview if there is one.
				if preview := m.getPreviewModel(); preview != nil {
					cmds = append(cmds, preview.Init())
				}
			} else {
				// toggle visibility = false for current model
				cmds = append(cmds, tui.ClearViewport())
			}
		case key.Matches(msg, localKeys.GrowPreview):
			// Grow the preview pane by shrinking the list pane
			m.userListHeightAdjustment--
			m.recalculateDimensions()
		case key.Matches(msg, localKeys.ShrinkPreview):
			// Shrink the preview pane by growing the list pane
			m.userListHeightAdjustment++
			m.recalculateDimensions()
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.Table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskKind, tui.WithParent(row.Value))
			}
		}
		if m.previewVisible && m.previewFocused {
			// Preview pane is visible and focused, so send key to current
			// preview model.
			cmds = append(cmds, m.updatePreviewModel(msg))
		} else {
			// Table pane is focused, so handle keys relevant to table rows.
			m.Table, cmd = m.Table.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.recalculateDimensions()
	default:
		// Forward remaining message types to both the table model and cached
		// resource models
		m.Table, cmd = m.Table.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.cache.UpdateAll(msg)...)
	}

	if m.previewVisible {
		// Get currently highlighted task and ensure a model exists for it, and
		// ensure that that model is the current model.
		if row, ok := m.Table.CurrentRow(); ok {
			model := m.cache.Get(row.Key)
			if model == nil {
				// Create model
				var err error
				model, err = m.maker.MakePreview(row.Value, m.width, m.previewHeight(), m.previewYPosition())
				if err != nil {
					return m, tui.ReportError(fmt.Errorf("making preview model: %w", err), "")
				}
				// Cache newly created model
				m.cache.Put(row.Key, model)
				// Initialize model
				cmds = append(cmds, model.Init())
			}
			if m.currentRowID != row.Key {
				// Current model has changed so it'll need re-initialising.
				cmds = append(cmds, model.Init())
			}
			m.currentRowID = row.Key
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model[R]) previewYPosition() int {
	return tui.DefaultYPosition + m.listHeight() + dividerHeight
}

func (m *Model[R]) recalculateDimensions() {
	m.Table, _ = m.Table.Update(tea.WindowSizeMsg{
		Height: m.listHeight(),
		Width:  m.width,
	})
	_ = m.cache.UpdateAll(tea.WindowSizeMsg{
		Height: m.previewHeight(),
		Width:  m.width,
	})
	_ = m.cache.UpdateAll(tui.YPositionMsg(m.previewYPosition()))
}

func (m Model[R]) listHeight() int {
	if m.previewVisible {
		// Ensure list pane is at least a height of 2 (the headings and one row)
		return max(2, defaultListPaneHeight+m.userListHeightAdjustment)
	}
	return m.height - dividerHeight
}

func (m Model[R]) previewHeight() int {
	// calculate height of preview pane after accounting for:
	// (a) height of list pane above
	// (b) height of dividing line in between panes.
	return max(0, m.height-m.listHeight()-dividerHeight)
}

func (m Model[R]) View() string {
	components := []string{m.Table.View()}
	// When preview pane is visible and there is a model cached for the
	// current row, then render the model's view in the pane.
	if m.previewVisible {
		if model := m.getPreviewModel(); model != nil {
			dividerChar := "▲"
			if m.previewFocused {
				dividerChar = "▼"
			}
			components = append(components,
				strings.Repeat(dividerChar, m.width),
				model.View(),
			)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Top, components...)
}

// getPreviewModel returns the model for the preview pane.
func (m Model[R]) getPreviewModel() tea.Model {
	row, ok := m.Table.CurrentRow()
	if !ok {
		return nil
	}
	return m.cache.Get(row.Key)
}

// updatePreviewModel updates the model for the preview pane.
func (m Model[R]) updatePreviewModel(msg tea.Msg) tea.Cmd {
	row, ok := m.Table.CurrentRow()
	if !ok {
		return nil
	}
	return m.cache.Update(row.Key, msg)
}
