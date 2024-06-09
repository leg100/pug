package split

import (
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
	// total width of borders to the left and right of a pane
	totalPaneBorderWidth = 2
	// total height of borders above and below a pane
	totalPaneBorderHeight = 2
)

type Options[R resource.Resource] struct {
	Columns      []table.Column
	Renderer     table.RowRenderer[R]
	TableOptions []table.Option[resource.ID, R]
	Width        int
	Height       int
	Maker        tui.Maker
}

func New[R resource.Resource](opts Options[R]) Model[R] {
	m := Model[R]{
		width:          opts.Width,
		height:         opts.Height,
		maker:          opts.Maker,
		cache:          tui.NewCache(),
		previewVisible: true,
	}
	// Create table for the top list pane
	m.Table = table.New(
		opts.Columns,
		opts.Renderer,
		m.paneWidth(),
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
	maker tui.Maker

	previewVisible bool
	previewFocused bool
	height         int
	width          int

	// userListHeightAdjustment is the adjustment the user has requested to the
	// default height of the list pane.
	userListHeightAdjustment int

	// cache of models for the previews
	cache *tui.Cache
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
		case key.Matches(msg, localKeys.TogglePreview):
			m.previewVisible = !m.previewVisible
			m.recalculateDimensions()
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
			// Preview pane is visible and focused, so send keys to the task
			// model for the currently highlighted table row if there is one.
			row, ok := m.Table.CurrentRow()
			if !ok {
				break
			}
			page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
			cmd := m.cache.Update(tui.NewCacheKey(page), msg)
			cmds = append(cmds, cmd)
		} else {
			// Table pane is focused, so handle keys relevant to table rows.
			//
			// TODO: when preview is focused, we also want these keys to be
			// handled for the current row (but not selected rows).
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
			page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
			if !m.cache.Exists(page) {
				// Create model
				model, err := m.maker.Make(row.Value, m.paneWidth(), m.previewHeight())
				if err != nil {
					return m, tui.ReportError(err, "making model for preview")
				}
				// Cache newly created model
				m.cache.Put(page, model)
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
		Width:  m.paneWidth(),
	})
	_ = m.cache.UpdateAll(tea.WindowSizeMsg{
		Height: m.previewHeight(),
		Width:  m.paneWidth(),
	})
}

func (m Model[R]) paneWidth() int {
	return m.width - totalPaneBorderWidth
}

func (m Model[R]) listHeight() int {
	if m.previewVisible {
		// Ensure list pane is at least a height of 2 (the headings and one row)
		return max(2, defaultListPaneHeight+m.userListHeightAdjustment)
	}
	return m.height - totalPaneBorderHeight
}

func (m Model[R]) previewHeight() int {
	// calculate height of preview pane after accounting for:
	// (a) height of list pane above
	// (b) height of borders above and below both panes
	return max(0, m.height-m.listHeight()-(totalPaneBorderHeight*2))
}

var (
	singlePaneBorder   = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	activePaneBorder   = lipgloss.NewStyle().Border(lipgloss.ThickBorder())
	inactivePaneBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(tui.LighterGrey)
)

func (m Model[R]) View() string {
	var (
		tableBorder   lipgloss.Style
		previewBorder lipgloss.Style
	)
	if !m.previewVisible {
		tableBorder = singlePaneBorder
	} else if m.previewFocused {
		tableBorder = inactivePaneBorder
		previewBorder = activePaneBorder
	} else {
		tableBorder = activePaneBorder
		previewBorder = inactivePaneBorder
	}
	components := []string{
		tableBorder.Render(m.Table.View()),
	}
	// When preview pane is visible and there is a model cached for the
	// current row, then render the model's view in the pane.
	if m.previewVisible {
		if model, ok := m.getPreviewModel(); ok {
			components = append(components, previewBorder.Render(model.View()))
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
	page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
	model := m.cache.Get(page)
	return model, model != nil
}
