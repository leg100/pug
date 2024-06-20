package table

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/go-runewidth"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"golang.org/x/exp/maps"
)

const (
	// Height of the table header
	headerHeight = 1
	// Height of filter widget
	filterHeight = 2
)

// Model defines a state for the table widget.
type Model[V resource.Resource] struct {
	cols        []Column
	rows        []Row[V]
	rowRenderer RowRenderer[V]
	focus       bool

	border      lipgloss.Border
	borderColor lipgloss.TerminalColor

	cursorRow       int
	cursorID        resource.ID
	userMovedCursor bool

	items    map[resource.ID]V
	sortFunc SortFunc[V]

	Selected   map[resource.ID]V
	selectable bool

	filter textinput.Model

	viewport viewport.Model
	start    int
	end      int

	// dimensions calcs
	width  int
	height int

	parent resource.Resource
}

// Column defines the table structure.
type Column struct {
	Key ColumnKey
	// TODO: Default to upper case of key
	Title          string
	Width          int
	FlexFactor     int
	TruncationFunc func(s string, w int, tail string) string
}

type ColumnKey string

type Row[V any] struct {
	ID    resource.ID
	Value V
}

type RowRenderer[V any] func(V) RenderedRow

// RenderedRow provides the rendered string for each column in a row.
type RenderedRow map[ColumnKey]string

type SortFunc[V any] func(V, V) int

// New creates a new model for the table widget.
func New[V resource.Resource](columns []Column, fn RowRenderer[V], width, height int, opts ...Option[V]) Model[V] {
	filter := textinput.New()
	filter.Prompt = "Filter: "

	m := Model[V]{
		viewport:    viewport.New(0, 0),
		rowRenderer: fn,
		items:       make(map[resource.ID]V),
		Selected:    make(map[resource.ID]V),
		selectable:  true,
		focus:       true,
		filter:      filter,
		border:      lipgloss.NormalBorder(),
	}
	for _, fn := range opts {
		fn(&m)
	}

	// Deliberately use range to copy column structs onto receiver, because the
	// caller may be using columns in multiple tables and columns are modified
	// by each table.
	//
	// TODO: use copy, which is more explicit
	for _, col := range columns {
		// Set default truncation function if unset
		if col.TruncationFunc == nil {
			col.TruncationFunc = defaultTruncationFunc
		}
		m.cols = append(m.cols, col)
	}

	m.setDimensions(width, height)

	return m
}

type Option[V resource.Resource] func(m *Model[V])

// WithSortFunc configures the table to sort rows using the given func.
func WithSortFunc[V resource.Resource](sortFunc func(V, V) int) Option[V] {
	return func(m *Model[V]) {
		m.sortFunc = sortFunc
	}
}

// WithSelectable sets whether rows are selectable.
func WithSelectable[V resource.Resource](s bool) Option[V] {
	return func(m *Model[V]) {
		m.selectable = s
	}
}

func WithParent[V resource.Resource](parent resource.Resource) Option[V] {
	return func(m *Model[V]) {
		m.parent = parent
	}
}

func (m *Model[V]) filterVisible() bool {
	// Filter is visible if it's either in focus, or it has a non-empty value.
	return m.filter.Focused() || m.filter.Value() != ""
}

// setDimensions sets the dimensions of the table.
func (m *Model[V]) setDimensions(width, height int) {
	m.height = height
	m.width = width

	// Accommodate height of table header and borders
	m.viewport.Height = max(0, height-headerHeight-2)
	if m.filterVisible() {
		// Accommodate height of filter widget
		m.viewport.Height = max(0, m.viewport.Height-filterHeight)
	}

	// Set available width for table to expand into, whilst respecting a
	// minimum width of 80, and accomodating border.
	m.viewport.Width = max(0, width-2)
	m.recalculateWidth()

	// TODO: should this always be called?
	m.UpdateViewport()
}

// Update is the Bubble Tea update loop.
func (m Model[V]) Update(msg tea.Msg) (Model[V], tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Navigation.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, keys.Navigation.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keys.Navigation.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, keys.Navigation.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, keys.Navigation.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, keys.Navigation.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, keys.Navigation.GotoTop):
			m.GotoTop()
		case key.Matches(msg, keys.Navigation.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, keys.Global.Select):
			m.ToggleSelection()
		case key.Matches(msg, keys.Global.SelectAll):
			m.SelectAll()
		case key.Matches(msg, keys.Global.SelectClear):
			m.DeselectAll()
		case key.Matches(msg, keys.Global.SelectRange):
			m.SelectRange()
		}
	case BulkInsertMsg[V]:
		for _, ws := range msg {
			m.items[ws.GetID()] = ws
		}
		m.SetItems(m.items)
	case resource.Event[V]:
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			m.items[msg.Payload.GetID()] = msg.Payload
			m.SetItems(m.items)
		case resource.DeletedEvent:
			delete(m.items, msg.Payload.GetID())
			m.SetItems(m.items)
		}
	case tea.WindowSizeMsg:
		m.setDimensions(msg.Width, msg.Height)
	case spinner.TickMsg:
		// Rows can contain spinners, so we re-render them whenever a tick is
		// received.
		m.UpdateViewport()
	case tui.FilterFocusReqMsg:
		// Focus the filter widget
		blink := m.filter.Focus()
		// Resize the viewport to accommodate the filter widget
		m.setDimensions(m.width, m.height)
		// Acknowledge the request, and start blinking the cursor.
		return m, tea.Batch(tui.CmdHandler(tui.FilterFocusAckMsg{}), blink)
	case tui.FilterBlurMsg:
		// Blur the filter widget
		m.filter.Blur()
		return m, nil
	case tui.FilterCloseMsg:
		// Close the filter widget
		m.filter.Blur()
		m.filter.SetValue("")
		// Unfilter table items
		m.SetItems(m.items)
		// Resize the viewport to take up the space now unoccupied
		m.setDimensions(m.width, m.height)
		return m, nil
	case tui.FilterKeyMsg:
		// unwrap key and send to filter widget
		kmsg := tea.KeyMsg(msg)
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(kmsg)
		// Filter table items
		m.SetItems(m.items)
		return m, cmd
	default:
		// Send any other messages to the filter if it is focused.
		if m.filter.Focused() {
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// Focused returns the focus state of the table.
func (m Model[V]) Focused() bool {
	return m.focus
}

// Focus focuses the table, allowing the user to move around the rows and
// interact.
func (m *Model[V]) Focus() {
	m.focus = true
	m.UpdateViewport()
}

// Blur blurs the table, preventing selection or movement.
func (m *Model[V]) Blur() {
	m.focus = false
	m.UpdateViewport()
}

// View renders the component.
func (m Model[V]) View() string {
	components := make([]string, 0, 3)
	if m.filterVisible() {
		components = append(components, tui.Regular.Margin(0, 1).Render(m.filter.View()))
		// Subtract 2 to accommodate border
		components = append(components, strings.Repeat("─", m.width-2))
	}
	components = append(components, m.headersView())
	components = append(components, m.viewport.View())
	content := lipgloss.JoinVertical(lipgloss.Top, components...)

	metadata := m.TotalString()

	// total length of top border runes, not including corners
	topBorderLength := max(0, m.width-lipgloss.Width(metadata)-2)
	topBorderLeftLength := topBorderLength / 2
	topBorderRightLength := topBorderLength - topBorderLeftLength

	topBorder := lipgloss.NewStyle().Foreground(m.borderColor).Render(fmt.Sprintf("%s%s%s%s%s", m.border.TopLeft, strings.Repeat(m.border.Top, topBorderLeftLength), metadata, strings.Repeat(m.border.Top, topBorderRightLength), m.border.TopRight))

	return lipgloss.JoinVertical(lipgloss.Top,
		topBorder,
		lipgloss.NewStyle().Border(m.border, false, true, true, true).BorderForeground(m.borderColor).Render(content),
	)
}

func (m *Model[V]) SetBorderStyle(border lipgloss.Border, color lipgloss.TerminalColor) {
	m.border = border
	m.borderColor = color
}

// UpdateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model[V]) UpdateViewport() {
	renderedRows := make([]string, 0, len(m.rows))

	// Render only rows from: m.cursor-m.viewport.Height to: m.cursor+m.viewport.Height
	// Constant runtime, independent of number of rows in a table.
	// Limits the number of renderedRows to a maximum of 2*m.viewport.Height
	if m.cursorRow >= 0 {
		m.start = clamp(m.cursorRow-m.viewport.Height, 0, m.cursorRow)
	} else {
		m.start = 0
	}
	m.end = clamp(m.cursorRow+m.viewport.Height, m.cursorRow, len(m.rows))
	for i := m.start; i < m.end; i++ {
		renderedRows = append(renderedRows, m.renderRow(i))
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

// CurrentRow returns the row on which the cursor currently sits. If the cursor
// is out of bounds then false is returned along with an empty row.
func (m Model[V]) CurrentRow() (Row[V], bool) {
	if m.cursorRow < 0 || m.cursorRow >= len(m.rows) {
		return *new(Row[V]), false
	}
	return m.rows[m.cursorRow], true
}

// SelectedOrCurrent returns either the selected rows, or if there are no
// selections, the current row
func (m Model[V]) SelectedOrCurrent() []Row[V] {
	if len(m.Selected) > 0 {
		rows := make([]Row[V], len(m.Selected))
		var i int
		for k, v := range m.Selected {
			rows[i] = Row[V]{ID: k, Value: v}
			i++
		}
		return rows
	}
	if row, ok := m.CurrentRow(); ok {
		return []Row[V]{row}
	}
	return nil
}

func (m Model[V]) SelectedOrCurrentIDs() []resource.ID {
	if len(m.Selected) > 0 {
		return maps.Keys(m.Selected)
	}
	if row, ok := m.CurrentRow(); ok {
		return []resource.ID{row.ID}
	}
	return nil
}

// ToggleSelection toggles the selection of the current row.
func (m *Model[V]) ToggleSelection() {
	if !m.selectable {
		return
	}
	current, ok := m.CurrentRow()
	if !ok {
		return
	}
	if _, isSelected := m.Selected[current.ID]; isSelected {
		delete(m.Selected, current.ID)
	} else {
		m.Selected[current.ID] = current.Value
	}
	m.UpdateViewport()
}

// ToggleSelectionByID toggles the selection of the row with the given id. If
// the id does not exist no action is taken.
func (m *Model[V]) ToggleSelectionByID(id resource.ID) {
	if !m.selectable {
		return
	}
	v, ok := m.items[id]
	if !ok {
		return
	}
	if _, isSelected := m.Selected[id]; isSelected {
		delete(m.Selected, id)
	} else {
		m.Selected[id] = v
	}
	m.UpdateViewport()
}

// SelectAll selects all rows. Any rows not currently selected are selected.
func (m *Model[V]) SelectAll() {
	if !m.selectable {
		return
	}

	for _, row := range m.rows {
		m.Selected[row.ID] = row.Value
	}
	m.UpdateViewport()
}

// DeselectAll de-selects any rows that are currently selected
func (m *Model[V]) DeselectAll() {
	if !m.selectable {
		return
	}

	m.Selected = make(map[resource.ID]V)
	m.UpdateViewport()
}

// SelectRange selects a range of rows. If the current row is *below* a selected
// row then rows between them are selected, including the current row.
// Otherwise, if the current row is *above* a selected row then rows between
// them are selected, including the current row. If there are no selected rows
// then no action is taken.
func (m *Model[V]) SelectRange() {
	if !m.selectable {
		return
	}
	if len(m.Selected) == 0 {
		return
	}
	// Determine the first row to select, and the number of rows to select.
	first := -1
	n := 0
	for i, row := range m.rows {
		if i == m.cursorRow && first > -1 && first < m.cursorRow {
			// Select rows before and including cursor
			n = m.cursorRow - first + 1
			break
		}
		if _, ok := m.Selected[row.ID]; !ok {
			// Ignore unselected rows
			continue
		}
		if i > m.cursorRow {
			// Select rows including cursor and all rows up to but not including
			// next selected row
			first = m.cursorRow
			n = i - m.cursorRow
			break
		}
		// Start selecting rows after this currently selected row.
		first = i + 1
	}
	for _, row := range m.rows[first : first+n] {
		m.Selected[row.ID] = row.Value
	}
	m.UpdateViewport()
}

func (m Model[V]) TotalString() string {
	// Calculate the top and bottom visible row ordinal numbers, starting from 1
	top := m.TopRowIndex() + 1
	bottom := m.BottomRowIndex() + 1

	// Only print range of rows if there are any rows
	var prefix string
	if (bottom - top) > 0 {
		prefix = fmt.Sprintf("%d-%d of ", top, bottom)
	}

	if m.filterVisible() {
		return prefix + fmt.Sprintf("%d/%d", len(m.rows), len(m.items))
	}
	return prefix + fmt.Sprintf("%d", len(m.rows))
}

// TopRowNumber returns the index of the top visible row.
func (m Model[V]) TopRowIndex() int {
	return m.start + max(0, m.viewport.YOffset)
}

// BottomRowNumber returns the index of the bottom visible row.
func (m Model[V]) BottomRowIndex() int {
	return clamp(m.start+m.viewport.YOffset+m.viewport.Height-1, m.TopRowIndex(), clamp(len(m.rows)-1, 0, len(m.rows)))
}

// SetItems sets new items on the table, overwriting existing items. If the
// table has a parent resource, then items that are not a descendent of that
// resource are skipped.
func (m *Model[V]) SetItems(items map[resource.ID]V) {
	// Skip non-descendent items
	if m.parent != nil {
		for k, v := range items {
			if !v.HasAncestor(m.parent.GetID()) {
				delete(items, k)
			}
		}
	}

	// Overwrite existing items
	m.items = items

	// Carry over existing selections.
	selections := make(map[resource.ID]V)

	// Overwrite existing rows
	m.rows = make([]Row[V], 0, len(items))

	// Convert items into rows, and carry across matching selections
	for id, it := range items {
		if m.filter.Value() != "" {
			// Filter rows using row renderer. If the filter value is a
			// substring of one of the rendered cells then add row. Otherwise,
			// skip row.
			//
			// NOTE: it is highly inefficient to render every row, every time
			// the user edits the filter value, particularly as the row renderer
			// can make data lookups on each invocation. But there is no obvious
			// alternative at present.
			filterMatch := func(v V) bool {
				for _, v := range m.rowRenderer(it) {
					// Remove ANSI escapes code before filtering
					v = internal.StripAnsi(v)
					if strings.Contains(v, m.filter.Value()) {
						return true
					}
				}
				return false
			}
			if !filterMatch(it) {
				// Skip item not matching filter
				continue
			}
		}
		m.rows = append(m.rows, Row[V]{ID: id, Value: it})
		if m.selectable {
			if _, ok := m.Selected[id]; ok {
				selections[id] = it
			}
		}
	}

	// Sort rows in-place
	if m.sortFunc != nil {
		slices.SortFunc(m.rows, func(i, j Row[V]) int {
			return m.sortFunc(i.Value, j.Value)
		})
	}

	// Overwrite existing selections, removing any that no longer have a
	// corresponding item.
	m.Selected = selections

	// Track item corresponding to the current cursor, but only if the user has
	// moved the cursor. Otherwise keep cursor on the first row.
	m.cursorRow = -1
	if m.userMovedCursor {
		for i, item := range m.rows {
			if item.ID == m.cursorID {
				// Found item corresponding to cursor, update its position
				m.cursorRow = i
			}
		}
	}
	// Check if item corresponding to cursor doesn't exist, which occurs when
	// items are removed, or the very first time the table is populated. If so,
	// set cursor to the first row.
	if m.cursorRow == -1 {
		m.cursorRow = 0
		if len(m.rows) > 0 {
			m.cursorID = m.rows[m.cursorRow].ID
		}
	}

	if m.cursorRow < m.TopRowIndex() {
		// Cursor is above visible lines; move cursor back into view
		m.MoveUp(0)
	} else if m.cursorRow > m.BottomRowIndex() {
		// Cursor is below visible lines; move cursor back into view
		m.MoveDown(0)
	} else {
		m.UpdateViewport()
	}
}

// MoveUp moves the current row up by any number of rows.
// It can not go above the first row.
func (m *Model[V]) MoveUp(n int) {
	m.moveCursor(-n)

	switch {
	case m.start == 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset, 0, m.cursorRow))
	case m.start < m.viewport.Height:
		m.viewport.YOffset = (clamp(clamp(m.viewport.YOffset+n, 0, m.cursorRow), 0, m.viewport.Height))
	case m.viewport.YOffset >= 1:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset+n, 1, m.viewport.Height))
	}

	m.UpdateViewport()
}

// MoveDown moves the current row down by any number of rows.
// It can not go below the last row.
func (m *Model[V]) MoveDown(n int) {
	m.moveCursor(n)
	m.UpdateViewport()

	switch {
	case m.end == len(m.rows) && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.viewport.Height))
	case m.cursorRow > (m.end-m.start)/2 && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.cursorRow))
	case m.viewport.YOffset > 1:
	case m.cursorRow > m.viewport.YOffset+m.viewport.Height-1:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset+1, 0, 1))
	}
}

func (m *Model[V]) moveCursor(n int) {
	m.cursorRow = clamp(m.cursorRow+n, 0, len(m.rows)-1)
	if len(m.rows) > 0 {
		m.cursorID = m.rows[m.cursorRow].ID
	}
	if n != 0 {
		m.userMovedCursor = true
	}
}

// GotoTop makes the top row the current row.
func (m *Model[V]) GotoTop() {
	m.MoveUp(m.cursorRow)
}

// GotoBottom makes the bottom row the current row.
func (m *Model[V]) GotoBottom() {
	m.MoveDown(len(m.rows))
}

func (m Model[V]) headersView() string {
	var s = make([]string, 0, len(m.cols))
	for _, col := range m.cols {
		style := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true)
		renderedCell := style.Render(runewidth.Truncate(col.Title, col.Width, "…"))
		s = append(s, tui.Regular.Copy().Padding(0, 1).Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, s...)
}

func (m *Model[V]) renderRow(rowIdx int) string {
	row := m.rows[rowIdx]

	var (
		background lipgloss.Color
		foreground lipgloss.Color
		current    bool
		selected   bool
	)
	if _, ok := m.Selected[row.ID]; ok {
		selected = true
	}
	if rowIdx == m.cursorRow {
		current = true
	}
	if current && selected {
		background = tui.CurrentAndSelectedBackground
		foreground = tui.CurrentAndSelectedForeground
	} else if current {
		background = tui.CurrentBackground
		foreground = tui.CurrentForeground
	} else if selected {
		background = tui.SelectedBackground
		foreground = tui.SelectedForeground
	}

	var renderedCells = make([]string, len(m.cols))
	cells := m.rowRenderer(row.Value)
	for i, col := range m.cols {
		content := cells[col.Key]
		// Truncate content if it is wider than column
		truncated := col.TruncationFunc(content, col.Width, "…")
		// Ensure content is all on one line.
		inlined := lipgloss.NewStyle().
			Width(col.Width).
			MaxWidth(col.Width).
			Inline(true).
			Render(truncated)
		// Apply block-styling to content
		boxed := lipgloss.NewStyle().
			Padding(0, 1).
			Render(inlined)
		renderedCells[i] = boxed
	}

	// Join cells together to form a row
	renderedRow := lipgloss.JoinHorizontal(lipgloss.Left, renderedCells...)

	// If current row or seleted rows, strip colors and apply background color
	if current || selected {
		renderedRow = internal.StripAnsi(renderedRow)
		renderedRow = lipgloss.NewStyle().
			Foreground(foreground).
			Background(background).
			Render(renderedRow)
	}
	return renderedRow
}

// Prune invokes the provided function with each selected value, and if the
// function returns true then it is de-selected. If there are any de-selections
// then an error is returned. If no pruning occurs then the id from each
// function invocation is returned.
//
// In the case where there are no selections then the current value is passed to
// the function, and if the function returns true then an error is reported. If
// it returns false then the resulting id is returned.
//
// If there are no rows in the table then a nil error is returned.
func (m *Model[V]) Prune(fn func(value V) (resource.ID, bool)) ([]resource.ID, error) {
	rows := m.SelectedOrCurrent()
	switch len(rows) {
	case 0:
		return nil, errors.New("no rows in table")
	case 1:
		// current row, no selections
		id, prune := fn(rows[0].Value)
		if prune {
			// the single current row is to be pruned, so report this as an
			// error
			return nil, fmt.Errorf("action is not applicable to the current row")
		}
		return []resource.ID{id}, nil
	default:
		// one or more selections: iterate thru and prune accordingly.
		var (
			ids    []resource.ID
			before = len(m.Selected)
			pruned int
		)
		for k, v := range m.Selected {
			id, prune := fn(v)
			if prune {
				// De-select
				m.ToggleSelectionByID(k)
				pruned++
				continue
			}
			ids = append(ids, id)
		}
		switch {
		case len(ids) == 0:
			return nil, errors.New("no selected rows are applicable to the given action")
		case len(ids) != before:
			// some rows have been pruned
			return nil, fmt.Errorf("de-selected %d inapplicable rows out of %d", pruned, before)
		}
		return ids, nil
	}
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
