package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/tui/common"
)

//type helpBindingFunc func(short bool, current common.Page) []key.Binding
//
//var helpBindingFuncs []helpBindingFunc
//
//func registerHelpBindings(f helpBindingFunc) {
//	helpBindingFuncs = append(helpBindingFuncs, f)
//}
//
//func getBindings(short bool, current common.Page) (bindings []key.Binding) {
//	for _, f := range helpBindingFuncs {
//		bindings = append(bindings, f(short, current)...)
//	}
//	if !short {
//		bindings = append(bindings, common.Keys.Quit)
//		bindings = append(bindings, common.Keys.Escape)
//	}
//	return
//}

type helpModel struct {
	current []key.Binding
	height  int
}

func newHelpModel(current []key.Binding) helpModel {
	return helpModel{current: current}
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case common.ViewSizeMsg:
		m.height = msg.Height
	case tea.KeyMsg:
		if key.Matches(msg, common.Keys.CloseHelp, common.Keys.Escape) {
			// close help and return to last state
			return m, common.CmdHandler(common.ReturnLastMsg{})
		}
	}
	return m, nil
}

func (m helpModel) Title() string {
	return "help"
}

func (m helpModel) View() string {
	// full view of help bindings
	return renderHelp(m.current, m.height-2)
}

func (m helpModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

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
		// num of rows in *this* column
		colrows := min(rows, len(bindings)-i)
		keys := make([]string, colrows)
		descs := make([]string, colrows)
		for j := 0; j < colrows; j++ {
			keys[j] = bindings[i+j].Help().Key
			descs[j] = bindings[i+j].Help().Desc
		}
		colnum := (i / rows) * 2
		cols[colnum] = keyStyle.Render(strings.Join(keys, "\n"))
		cols[colnum+1] = descStyle.Render(strings.Join(descs, "\n"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
