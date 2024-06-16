package split

import (
	"fmt"

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
	defaultListPaneHeight = 15
	// previewVisibleDefault sets the default visibility for the preview pane.
	previewVisibleDefault = true
)

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
		previewVisible: previewVisibleDefault,
	}
	// Create table for the top list pane
	m.Table = table.New(
		opts.Columns,
		opts.Renderer,
		opts.Width,
		m.listHeight(),
		opts.TableOptions...,
	)
	m.setBorderStyles()
	return m
}

// Model is a composition of two models corresponding to two panes: a top pane
// is a list of resources; the bottom pane provides further details of the
// resource corresponding to the current row in the list - this pane is known as
// the 'preview'.
type Model[R resource.Resource] struct {
	Table table.Model[R]
	maker tui.Maker

	previewVisible bool
	previewFocused bool
	height         int
	width          int

	previewBorder      lipgloss.Border
	previewBorderColor lipgloss.Color

	// userListHeightAdjustment is the adjustment the user has requested to the
	// default height of the list pane.
	userListHeightAdjustment int

	// cache of models for the previews
	cache *cache
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
		case key.Matches(msg, keys.Navigation.SwitchPane):
			m.previewFocused = !m.previewFocused
			m.setBorderStyles()
		case key.Matches(msg, Keys.TogglePreview):
			m.previewVisible = !m.previewVisible
			m.setBorderStyles()
			m.recalculateDimensions()
		case key.Matches(msg, Keys.GrowPreview):
			// Grow the preview pane by shrinking the list pane
			m.userListHeightAdjustment--
			m.recalculateDimensions()
		case key.Matches(msg, Keys.ShrinkPreview):
			// Shrink the preview pane by growing the list pane
			m.userListHeightAdjustment++
			m.recalculateDimensions()
		}
		if m.previewVisible && m.previewFocused {
			// Preview pane is visible and focused, so send keys to the preview
			// model for the currently highlighted table row if there is one.
			row, ok := m.Table.CurrentRow()
			if !ok {
				break
			}
			cmd := m.cache.Update(row.ID, msg)
			cmds = append(cmds, cmd)
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
		// Get current table row and ensure a model exists for it, and
		// ensure that that model is the current model.
		if row, ok := m.Table.CurrentRow(); ok {
			if model := m.cache.Get(row.ID); model == nil {
				// Create model
				model, err := m.maker.Make(row.Value, m.previewWidth(), m.previewHeight())
				if err != nil {
					return m, tui.ReportError(fmt.Errorf("making model for preview: %w", err))
				}
				// Cache newly created model
				m.cache.Put(row.ID, model)
				// Set border style on model
				m.setBorderStyles()
				// Initialize model
				cmds = append(cmds, model.Init())
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model[R]) recalculateDimensions() {
	m.Table, _ = m.Table.Update(tea.WindowSizeMsg{
		Height: m.listHeight(),
		Width:  m.width,
	})
	_ = m.cache.UpdateAll(tea.WindowSizeMsg{
		Height: m.previewHeight(),
		Width:  m.previewWidth(),
	})
}

func (m Model[R]) previewWidth() int {
	// Subtract 2 to accommodate borders
	return m.width - 2
}

func (m Model[R]) listHeight() int {
	if m.previewVisible {
		// List height cannot exceed available height - 3 is the min height of
		// the preview pane including borders.
		return min(defaultListPaneHeight+m.userListHeightAdjustment, m.height-3)
	}
	return m.height
}

// previewHeight returns the height of the preview pane, not including borders
func (m Model[R]) previewHeight() int {
	// Calculate height of preview pane after accommodating list and borders.
	return max(0, m.height-m.listHeight()-2)
}

func (m *Model[R]) setBorderStyles() {
	if m.previewVisible {
		if m.previewFocused {
			m.Table.SetBorderStyle(lipgloss.NormalBorder(), tui.LighterGrey)
			m.previewBorder = lipgloss.ThickBorder()
			m.previewBorderColor = tui.Blue
		} else {
			m.Table.SetBorderStyle(lipgloss.ThickBorder(), tui.Black)
			m.previewBorder = lipgloss.NormalBorder()
			m.previewBorderColor = tui.LighterGrey
		}
	} else {
		m.Table.SetBorderStyle(lipgloss.NormalBorder(), tui.Black)
	}
}

func (m Model[R]) View() string {
	components := []string{m.Table.View()}
	// When preview pane is visible and there is a model cached for the
	// current row, then render the model's view in the pane.
	if m.previewVisible {
		if model, ok := m.getPreviewModel(); ok {
			style := lipgloss.NewStyle().
				Border(m.previewBorder).
				BorderForeground(m.previewBorderColor)
			components = append(components, style.Render(model.View()))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Top, components...)
}

// getPreviewModel returns the model for the preview pane.
func (m Model[R]) getPreviewModel() (tea.Model, bool) {
	row, ok := m.Table.CurrentRow()
	if !ok {
		return nil, false
	}
	model := m.cache.Get(row.ID)
	return model, model != nil
}
