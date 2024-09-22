package table

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/go-runewidth"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"golang.org/x/exp/maps"
)

const (
	// Height of the table header
	headerHeight = 1
	// Height of filter widget
	filterHeight = 2
	// Minimum recommended height for the table widget. Respecting this minimum
	// ensures the header and the borders and the filter widget are visible.
	MinHeight = 6
)

// Model defines a state for the table widget.
type Model[V resource.Resource] struct {
	cols        []Column
	rows        []Row[V]
	rowRenderer RowRenderer[V]
	focus       bool
	rendered    map[resource.ID]RenderedRow

	border      lipgloss.Border
	borderColor lipgloss.TerminalColor

	currentRowIndex int
	currentRowID    resource.ID

	// items are the unfiltered set of items available to the table.
	items    map[resource.ID]V
	sortFunc SortFunc[V]

	selected   map[resource.ID]V
	selectable bool

	filter textinput.Model

	// index of first visible row
	start int

	// width of table without borders
	width int
	// height of table without borders
	height int
}

// Column defines the table structure.
type Column struct {
	Key ColumnKey
	// TODO: Default to upper case of key
	Title          string
	Width          int
	FlexFactor     int
	TruncationFunc func(s string, w int, tail string) string
	// RightAlign aligns content to the right. If false, content is aligned to
	// the left.
	RightAlign bool
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
func New[V resource.Resource](cols []Column, fn RowRenderer[V], width, height int, opts ...Option[V]) Model[V] {
	filter := textinput.New()
	filter.Prompt = "Filter: "

	m := Model[V]{
		rowRenderer:     fn,
		items:           make(map[resource.ID]V),
		rendered:        make(map[resource.ID]RenderedRow),
		selected:        make(map[resource.ID]V),
		selectable:      true,
		focus:           true,
		filter:          filter,
		border:          lipgloss.NormalBorder(),
		currentRowIndex: -1,
	}
	for _, fn := range opts {
		fn(&m)
	}

	// Copy column structs onto receiver, because the caller may modify columns.
	m.cols = make([]Column, len(cols))
	copy(m.cols, cols)
	// For each column, set default truncation function if unset.
	for i, col := range m.cols {
		if col.TruncationFunc == nil {
			m.cols[i].TruncationFunc = defaultTruncationFunc
		}
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

func (m *Model[V]) filterVisible() bool {
	// Filter is visible if it's either in focus, or it has a non-empty value.
	return m.filter.Focused() || m.filter.Value() != ""
}

// setDimensions sets the dimensions of the table.
func (m *Model[V]) setDimensions(width, height int) {
	// Adjust height to accomodate borders
	m.height = height - 2
	// Adjust width to accomodate borders
	m.width = width - 2
	m.setColumnWidths()

	m.setStart()
}

// rowAreaHeight returns the height of the terminal allocated to rows.
func (m Model[V]) rowAreaHeight() int {
	height := max(0, m.height-headerHeight)

	if m.filterVisible() {
		// Accommodate height of filter widget
		return max(0, height-filterHeight)
	}
	return height
}

// visibleRows returns the number of renderable visible rows.
func (m Model[V]) visibleRows() int {
	// The number of visible rows cannot exceed the row area height.
	return min(m.rowAreaHeight(), len(m.rows)-m.start)
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
			m.MoveUp(m.rowAreaHeight())
		case key.Matches(msg, keys.Navigation.PageDown):
			m.MoveDown(m.rowAreaHeight())
		case key.Matches(msg, keys.Navigation.HalfPageUp):
			m.MoveUp(m.rowAreaHeight() / 2)
		case key.Matches(msg, keys.Navigation.HalfPageDown):
			m.MoveDown(m.rowAreaHeight() / 2)
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
		m.AddItems(msg...)
	case resource.Event[V]:
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			m.AddItems(msg.Payload)
		case resource.DeletedEvent:
			m.removeItem(msg.Payload)
		}
	case tea.WindowSizeMsg:
		m.setDimensions(msg.Width, msg.Height)
	case tui.FilterFocusReqMsg:
		// Focus the filter widget
		blink := m.filter.Focus()
		// Start blinking the cursor.
		return m, blink
	case tui.FilterBlurMsg:
		// Blur the filter widget
		m.filter.Blur()
		return m, nil
	case tui.FilterCloseMsg:
		// Close the filter widget
		m.filter.Blur()
		m.filter.SetValue("")
		// Unfilter table items
		m.setRows(maps.Values(m.items)...)
		return m, nil
	case tui.FilterKeyMsg:
		// unwrap key and send to filter widget
		kmsg := tea.KeyMsg(msg)
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(kmsg)
		// Filter table items
		m.setRows(maps.Values(m.items)...)
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
}

// Blur blurs the table, preventing selection or movement.
func (m *Model[V]) Blur() {
	m.focus = false
}

// View renders the table.
func (m Model[V]) View() string {
	// Table is composed of a vertical stack of components:
	// (a) optional filter widget
	// (b) header
	// (c) rows + scrollbar
	components := make([]string, 0, 1+1+m.visibleRows())
	if m.filterVisible() {
		components = append(components, tui.Regular.Margin(0, 1).Render(m.filter.View()))
		// Add horizontal rule between filter widget and table
		components = append(components, strings.Repeat("─", m.width))
	}
	components = append(components, m.headersView())
	// Generate scrollbar
	scrollbar := tui.Scrollbar(m.rowAreaHeight(), len(m.rows), m.visibleRows(), m.start)
	// Get all the visible rows
	var rows []string
	for i := range m.visibleRows() {
		rows = append(rows, m.renderRow(m.start+i))
	}
	rowarea := lipgloss.NewStyle().Width(m.width - tui.ScrollbarWidth).Render(
		strings.Join(rows, "\n"),
	)
	// Put rows alongside the scrollbar to the right.
	components = append(components, lipgloss.JoinHorizontal(lipgloss.Top, rowarea, scrollbar))
	// Render table components, ensuring it is at least a min height
	content := lipgloss.NewStyle().
		Height(m.height).
		Render(lipgloss.JoinVertical(lipgloss.Top, components...))
	// Render table metadata
	var metadata string
	{
		// Calculate the top and bottom visible row ordinal numbers
		top := m.start + 1
		bottom := m.start + m.visibleRows()
		prefix := fmt.Sprintf("%d-%d of ", top, bottom)
		if m.filterVisible() {
			metadata = prefix + fmt.Sprintf("%d/%d", len(m.rows), len(m.items))
		} else {
			metadata = prefix + strconv.Itoa(len(m.rows))
		}
	}
	// Render top border with metadata in the center
	var topBorder string
	{
		// total length of top border runes, not including corners
		length := max(0, m.width-lipgloss.Width(metadata))
		leftLength := length / 2
		rightLength := max(0, length-leftLength)
		topBorder = lipgloss.JoinHorizontal(lipgloss.Left,
			m.border.TopLeft,
			strings.Repeat(m.border.Top, leftLength),
			metadata,
			strings.Repeat(m.border.Top, rightLength),
			m.border.TopRight,
		)
	}
	// Join top border with table components wrapped with borders on remaining
	// sides.
	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Foreground(m.borderColor).Render(topBorder),
		lipgloss.NewStyle().Border(m.border, false, true, true, true).BorderForeground(m.borderColor).Render(content),
	)
}

func (m *Model[V]) SetBorderStyle(border lipgloss.Border, color lipgloss.TerminalColor) {
	m.border = border
	m.borderColor = color
}

// CurrentRow returns the current row the user has highlighted. If its index is
// out of bounds then false is returned along with an empty row.
func (m Model[V]) CurrentRow() (Row[V], bool) {
	if m.currentRowIndex < 0 || m.currentRowIndex >= len(m.rows) {
		return *new(Row[V]), false
	}
	return m.rows[m.currentRowIndex], true
}

// SelectedOrCurrent returns either the selected rows, or if there are no
// selections, the current row
func (m Model[V]) SelectedOrCurrent() []Row[V] {
	if len(m.selected) > 0 {
		rows := make([]Row[V], len(m.selected))
		var i int
		for k, v := range m.selected {
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
	if len(m.selected) > 0 {
		return maps.Keys(m.selected)
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
	if _, isSelected := m.selected[current.ID]; isSelected {
		delete(m.selected, current.ID)
	} else {
		m.selected[current.ID] = current.Value
	}
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
	if _, isSelected := m.selected[id]; isSelected {
		delete(m.selected, id)
	} else {
		m.selected[id] = v
	}
}

// SelectAll selects all rows. Any rows not currently selected are selected.
func (m *Model[V]) SelectAll() {
	if !m.selectable {
		return
	}

	for _, row := range m.rows {
		m.selected[row.ID] = row.Value
	}
}

// DeselectAll de-selects any rows that are currently selected
func (m *Model[V]) DeselectAll() {
	if !m.selectable {
		return
	}

	m.selected = make(map[resource.ID]V)
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
	if len(m.selected) == 0 {
		return
	}
	// Determine the first row to select, and the number of rows to select.
	first := -1
	n := 0
	for i, row := range m.rows {
		if i == m.currentRowIndex && first > -1 && first < m.currentRowIndex {
			// Select rows before and including current row
			n = m.currentRowIndex - first + 1
			break
		}
		if _, ok := m.selected[row.ID]; !ok {
			// Ignore unselected rows
			continue
		}
		if i > m.currentRowIndex {
			// Select rows including current row and all rows up to but not
			// including next selected row
			first = m.currentRowIndex
			n = i - m.currentRowIndex
			break
		}
		// Start selecting rows after this currently selected row.
		first = i + 1
	}
	for _, row := range m.rows[first : first+n] {
		m.selected[row.ID] = row.Value
	}
}

// SetItems overwrites all existing items in the table with items.
func (m *Model[V]) SetItems(items ...V) {
	m.items = make(map[resource.ID]V)
	m.rendered = make(map[resource.ID]RenderedRow)
	m.AddItems(items...)
}

// AddItems idempotently adds items to the table, updating any items that exist
// on the table already.
func (m *Model[V]) AddItems(items ...V) {
	for _, item := range items {
		// Add/update item
		m.items[item.GetID()] = item
		// (Re-)render item's row.
		m.rendered[item.GetID()] = m.rowRenderer(item)
	}
	m.setRows(maps.Values(m.items)...)
}

func (m *Model[V]) removeItem(item V) {
	delete(m.rendered, item.GetID())
	delete(m.items, item.GetID())
	delete(m.selected, item.GetID())
	for i, row := range m.rows {
		if row.ID == item.GetID() {
			// TODO: this might well produce a memory leak. See note:
			// https://go.dev/wiki/SliceTricks#delete-without-preserving-order
			m.rows = append(m.rows[:i], m.rows[i+1:]...)
			break
		}
	}
	if item.GetID() == m.currentRowID {
		// If item being removed is the current row the make the row above it
		// the new current row. (MoveUp also calls setStart, see below).
		m.MoveUp(1)
	} else {
		// Removing item may well affect index of first visible row, so
		// re-calculate just in case.
		m.setStart()
	}
}

func (m *Model[V]) setRows(items ...V) {
	selected := make(map[resource.ID]V)
	m.rows = make([]Row[V], 0, len(items))
	for _, item := range items {
		if m.filterVisible() && !m.matchFilter(item.GetID()) {
			// Skip item that doesn't match filter
			continue
		}
		m.rows = append(m.rows, Row[V]{ID: item.GetID(), Value: item})
		if m.selectable {
			if _, ok := m.selected[item.GetID()]; ok {
				selected[item.GetID()] = item
			}
		}
	}
	m.selected = selected
	// Sort rows in-place
	if m.sortFunc != nil {
		slices.SortFunc(m.rows, func(i, j Row[V]) int {
			return m.sortFunc(i.Value, j.Value)
		})
	}
	// Track current row index
	m.currentRowIndex = -1
	for i, row := range m.rows {
		if row.ID == m.currentRowID {
			m.currentRowIndex = i
			break
		}
	}
	// Check if item corresponding to current row doesn't exist, which occurs
	// the very first time the table is populated. If so, set current row to the
	// first row.
	if len(m.rows) > 0 && m.currentRowIndex == -1 {
		m.currentRowIndex = 0
		m.currentRowID = m.rows[m.currentRowIndex].ID
	}
	m.setStart()
}

// matchFilter returns true if the item with the given ID matches the filter
// value.
func (m *Model[V]) matchFilter(id resource.ID) bool {
	for _, col := range m.rendered[id] {
		// Remove ANSI escapes code before filtering
		stripped := internal.StripAnsi(col)
		if strings.Contains(stripped, m.filter.Value()) {
			return true
		}
	}
	return false
}

// MoveUp moves the current row up by any number of rows.
// It can not go above the first row.
func (m *Model[V]) MoveUp(n int) {
	m.moveCurrentRow(-n)
}

// MoveDown moves the current row down by any number of rows.
// It can not go below the last row.
func (m *Model[V]) MoveDown(n int) {
	m.moveCurrentRow(n)
}

func (m *Model[V]) moveCurrentRow(n int) {
	if len(m.rows) > 0 {
		m.currentRowIndex = clamp(m.currentRowIndex+n, 0, len(m.rows)-1)
		m.currentRowID = m.rows[m.currentRowIndex].ID
		m.setStart()
	}
}

func (m *Model[V]) setStart() {
	// Start index must be at least the current row index minus the max number
	// of visible rows.
	minimum := max(0, m.currentRowIndex-m.rowAreaHeight()+1)
	// Start index must be at most the lesser of:
	// (a) the current row index, or
	// (b) the number of rows minus the maximum number of visible rows (as many
	// rows as possible are rendered)
	maximum := max(0, min(m.currentRowIndex, len(m.rows)-m.rowAreaHeight()))
	m.start = clamp(m.start, minimum, maximum)
}

// GotoTop makes the top row the current row.
func (m *Model[V]) GotoTop() {
	m.MoveUp(m.currentRowIndex)
}

// GotoBottom makes the bottom row the current row.
func (m *Model[V]) GotoBottom() {
	m.MoveDown(len(m.rows))
}

func (m Model[V]) headersView() string {
	var s = make([]string, 0, len(m.cols))
	for _, col := range m.cols {
		style := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true)
		if col.RightAlign {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		renderedCell := style.Render(runewidth.Truncate(col.Title, col.Width, "…"))
		s = append(s, tui.Regular.Padding(0, 1).Render(renderedCell))
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
	if _, ok := m.selected[row.ID]; ok {
		selected = true
	}
	if rowIdx == m.currentRowIndex {
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

	cells := m.rendered[row.ID]
	styledCells := make([]string, len(m.cols))
	for i, col := range m.cols {
		content := cells[col.Key]
		// Truncate content if it is wider than column
		truncated := col.TruncationFunc(content, col.Width, "…")
		// Ensure content is all on one line.
		style := lipgloss.NewStyle().
			Width(col.Width).
			MaxWidth(col.Width).
			Inline(true)
		if col.RightAlign {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		inlined := style.Render(truncated)
		// Apply block-styling to content
		boxed := lipgloss.NewStyle().
			Padding(0, 1).
			Render(inlined)
		styledCells[i] = boxed
	}

	// Join cells together to form a row
	renderedRow := lipgloss.JoinHorizontal(lipgloss.Left, styledCells...)

	// If current row or selected rows, strip colors and apply background color
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
func (m *Model[V]) Prune(fn func(value V) (task.Spec, error)) ([]task.Spec, error) {
	rows := m.SelectedOrCurrent()
	switch len(rows) {
	case 0:
		return nil, errors.New("no rows in table")
	case 1:
		// current row, no selections
		spec, err := fn(rows[0].Value)
		if err != nil {
			// the single current row is to be pruned, so report this as an
			// error
			return nil, err
		}
		return []task.Spec{spec}, nil
	default:
		// one or more selections: iterate thru and prune accordingly.
		var (
			ids    []task.Spec
			before = len(m.selected)
			pruned int
		)
		for k, v := range m.selected {
			spec, err := fn(v)
			if err != nil {
				// De-select
				m.ToggleSelectionByID(k)
				pruned++
				continue
			}
			ids = append(ids, spec)
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
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
