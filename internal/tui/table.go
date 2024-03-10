package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/common"
)

const (
	ColKeyID        = "id"
	ColKeyTime      = "time"
	ColKeyMessage   = "message"
	ColKeyLevel     = "level"
	ColKeyName      = "name"
	ColKeyModule    = "module"
	ColKeyWorkspace = "workspace"
	ColKeyStatus    = "status"
	ColKeyAgo       = "ago"
	ColKeyCommand   = "command"
	ColKeyData      = "data"
)

type tableModel[T tableItem] struct {
	table.Model

	parent    resource.Resource
	data      map[resource.ID]T
	rowMaker  func(T) table.RowData
	height    int
	width     int
	selectAll bool
}

type tableModelOptions[T tableItem] struct {
	parent            resource.Resource
	rowMaker          func(T) table.RowData
	columns           []table.Column
	disableSelections bool
}

type tableItem interface {
	resource.Entity
	HasAncestor(id resource.ID) bool
}

func newTableModel[T tableItem](opts tableModelOptions[T]) tableModel[T] {
	table := table.New(opts.columns).
		Focused(true).
		WithMultiline(false).
		SortByDesc(ColKeyTime).
		ThenSortByAsc(ColKeyModule).
		WithHeaderVisibility(true).
		WithFooterVisibility(false).
		HighlightStyle(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Background(lipgloss.Color("#a9aaab")),
		).
		// TODO: disable for logs
		SelectableRows(!opts.disableSelections).
		Border(
			table.Border{
				Top:    " ",
				Left:   " ",
				Right:  " ",
				Bottom: " ",

				TopRight:    " ",
				TopLeft:     " ",
				BottomRight: " ",
				BottomLeft:  " ",

				TopJunction:    " ",
				LeftJunction:   " ",
				RightJunction:  " ",
				BottomJunction: " ",
				InnerJunction:  " ",

				InnerDivider: " ",
			},
		).
		WithPaginationWrapping(false)
	return tableModel[T]{
		Model:    table,
		data:     make(map[resource.ID]T),
		rowMaker: opts.rowMaker,
	}
}

func (m tableModel[T]) Update(msg tea.Msg) (tableModel[T], tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case bulkInsertMsg[T]:
		for _, res := range msg {
			// Only add resource to table if the resource is a child of the
			// parent
			if !res.HasAncestor(m.parent.ID()) {
				return m, nil
			}
			m.data[res.ID()] = res
		}
		m.Model = m.WithRows(m.toRows())
	case resource.Event[T]:
		// Only add event resource to table if the resource is a child of the
		// parent
		if !msg.Payload.HasAncestor(m.parent.ID()) {
			return m, nil
		}
		switch msg.Type {
		case resource.CreatedEvent:
			m.data[msg.Payload.ID()] = msg.Payload
		case resource.UpdatedEvent:
			m.data[msg.Payload.ID()] = msg.Payload
		case resource.DeletedEvent:
			delete(m.data, msg.Payload.ID())
		}
		m.Model = m.WithRows(m.toRows())
	case common.ViewSizeMsg:
		m.Model = m.WithTargetWidth(msg.Width)
		m.Model = m.WithPageSize(msg.Height - tableMetaHeight)
		m.height = msg.Height
		m.width = msg.Width
	}
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

const (
	statusBarHeight = 1
	// Aggregate height of the table header and spacing between header and rows,
	// and space where the footer would be.
	tableMetaHeight = 5
)

func (m tableModel[T]) View() string {
	return lipgloss.NewStyle().
		Height(m.height - statusBarHeight).
		Render(m.Model.View())
}

func (m tableModel[T]) Pagination() string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#a8a7a5")).
		Foreground(White).
		Padding(0, 1).
		Margin(0, 1).
		Render(fmt.Sprintf("%d/%d", m.CurrentPage(), m.MaxPages()))
}

func (m tableModel[T]) toRows() []table.Row {
	to := make([]table.Row, len(m.data))
	var c int
	for id, res := range m.data {
		row := m.rowMaker(res)
		row[ColKeyData] = res
		row[ColKeyID] = id
		to[c] = table.NewRow(row)
		c++
	}
	return to
}

func (m *tableModel[T]) updateRows(fn func(table.Row) table.Row) {
	rows := make([]table.Row, len(m.data))
	for i, row := range m.toRows() {
		rows[i] = fn(row)
	}
	m.Model = m.WithRows(rows)
}

func (m tableModel[T]) highlighted() (bool, T) {
	if m.TotalRows() == 0 {
		return false, *new(T)
	}
	return true, m.HighlightedRow().Data[ColKeyData].(T)
}

func (m tableModel[T]) selected() []T {
	rows := m.SelectedRows()
	data := make([]T, len(rows))
	for i, row := range rows {
		data[i] = row.Data[ColKeyData].(T)
	}
	return data
}

func (m tableModel[T]) selectedIDs() []resource.ID {
	rows := m.SelectedRows()
	ids := make([]resource.ID, len(rows))
	for i, row := range rows {
		ids[i] = row.Data[ColKeyID].(resource.ID)
	}
	return ids
}

func (m tableModel[T]) highlightedOrSelected() []T {
	if selected := m.selected(); len(selected) > 0 {
		return selected
	}
	if ok, res := m.highlighted(); ok {
		return []T{res}
	}
	return nil
}

func (m tableModel[T]) highlightedOrSelectedIDs() []resource.ID {
	if selectedIDs := m.selectedIDs(); len(selectedIDs) > 0 {
		return selectedIDs
	}
	if ok, res := m.highlighted(); ok {
		return []resource.ID{res.ID()}
	}
	return nil
}

// bulkInsertMsg performs a bulk insertion of items into a table
type bulkInsertMsg[T any] []T

// deselectMsg deselects any table rows currently selected.
type deselectMsg struct{}
