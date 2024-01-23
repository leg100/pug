package internal

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

func renderShortHelp(bindings []key.Binding) string {
	if bindings == nil {
		return ""
	}
	bindings = append(
		[]key.Binding{keys.Quit, keys.Help},
		bindings...,
	)
	return renderHelp(bindings, 3)
}

func renderLongHelp(bindings []key.Binding, height int) string {
	bindings = append(
		[]key.Binding{keys.Quit, keys.CloseHelp},
		bindings...,
	)
	return renderHelp(bindings, height-2)
}

func renderHelp(bindings []key.Binding, rows int) string {
	var (
		// a column of keys and a column of descriptions for each group of three
		// bindings
		cols      = make([]string, 2*int(math.Ceil(float64(len(bindings))/float64(rows))))
		keyStyle  = regular.Bold(true).Align(lipgloss.Right).Margin(0, 1, 0, 2)
		descStyle = regular.Align(lipgloss.Left)
	)

	// iterate thru each column of keys/descs
	for i := 0; i < len(bindings); i += rows {
		// num of rows in *this* column
		rows := min(rows, len(bindings)-i)
		keys := make([]string, rows)
		descs := make([]string, rows)
		for j := 0; j < rows; j++ {
			keys[j] = bindings[i+j].Help().Key
			descs[j] = bindings[i+j].Help().Desc
		}
		colnum := (i / rows) * 2
		cols[colnum] = keyStyle.Render(strings.Join(keys, "\n"))
		cols[colnum+1] = descStyle.Render(strings.Join(descs, "\n"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
