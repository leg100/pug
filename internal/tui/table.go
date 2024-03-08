package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/common"
	"golang.org/x/exp/maps"
)

func newDefaultTable(cols ...table.Column) table.Model {
	return table.New(cols).
		Focused(true).
		WithMultiline(false).
		SortByDesc(common.ColKeyTime).
		ThenSortByAsc(common.ColKeyPath).
		WithHeaderVisibility(true).
		WithFooterVisibility(false).
		HighlightStyle(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Background(lipgloss.Color(DarkGrey)),
		).
		// TODO: disable for logs
		SelectableRows(true).
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
}

type tableModel[T fmt.Stringer] struct {
	table.Model

	data     map[resource.ID]T
	rowMaker func(T) table.RowData
}

func newTable[T fmt.Stringer](rowMaker func(T) table.RowData, cols ...table.Column) *tableModel[T] {
	return &tableModel[T]{
		Model:    newDefaultTable(cols...),
		data:     make(map[resource.ID]T),
		rowMaker: rowMaker,
	}
}

func (m tableModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.SelectAll):
			selected := make([]table.Row, len(m.data))
			for i, res := range maps.Values(m.data) {
				selected[i] = table.NewRow(m.rowMaker(res)).
					Selected(true)
			}
			m.Model = m.WithRows(selected)
		}
	}
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
