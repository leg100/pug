package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/tui/common"
)

func RenderShort(current []key.Binding) string {
	return renderHelp(current, 2)
}

func renderHelp(bindings []key.Binding, rows int) string {
	var (
		// 1 pair of columns per binding: 1 for keys, 1 for descriptions
		// the number of pairs of columns is determined by the number of
		// bindings and the number of rows.
		cols      = make([]string, 2*int(math.Ceil(float64(len(bindings))/float64(rows))))
		keyStyle  = common.Regular.Copy().Bold(true).Align(lipgloss.Right).Margin(0, 1, 0, 2)
		descStyle = common.Regular.Copy().Align(lipgloss.Left)
	)

	// iterate thru each pair of columns of keys/descs
	for i := 0; i < len(bindings); i += rows {
		// num of cells in *this* column
		cells := min(rows, len(bindings)-i)
		keys := make([]string, cells)
		descs := make([]string, cells)
		for j := 0; j < cells; j++ {
			keys[j] = bindings[i+j].Help().Key
			descs[j] = bindings[i+j].Help().Desc
		}
		colnum := (i / rows) * 2
		cols[colnum] = keyStyle.Render(strings.Join(keys, "\n"))
		cols[colnum+1] = descStyle.Render(strings.Join(descs, "\n"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
