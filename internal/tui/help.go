package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpBindingFunc func(short bool, current State) []key.Binding

var helpBindingFuncs []helpBindingFunc

func registerHelpBindings(f helpBindingFunc) {
	helpBindingFuncs = append(helpBindingFuncs, f)
}

func getBindings(short bool, current State) (bindings []key.Binding) {
	for _, f := range helpBindingFuncs {
		bindings = append(bindings, f(short, current)...)
	}
	if !short {
		bindings = append(bindings, Keys.Quit)
		bindings = append(bindings, Keys.Escape)
	}
	return
}

const helpState State = "help"

type help struct {
	current, last State
	height        int
}

func newHelp(current State) help {
	return help{current: current}
}

func (m help) Init() tea.Cmd {
	return nil
}

func (m help) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ViewSizeMsg:
		m.height = msg.Height
	case ChangeStateMsg:
		m.current = msg.To
	case GlobalKeyMsg:
		if key.Matches(msg.KeyMsg, Keys.Help) {
			if msg.Current != helpState {
				// open help, keeping reference to last state
				m.last = m.current
				return m, ChangeState(helpState)
			}
		}
	case tea.KeyMsg:
		if key.Matches(msg, Keys.CloseHelp, Keys.Escape) {
			// close help and return to last state
			return m, ChangeState(m.last)
		}
	}
	return m, nil
}

func (m help) Title() string {
	return "help"
}

func (m help) View() string {
	// full view of help bindings
	return renderHelp(getBindings(false, m.last), m.height-2)
}

func RenderShort(current State) string {
	if current == helpState {
		return renderHelp([]key.Binding{Keys.CloseHelp}, 1)
	}
	return renderHelp(getBindings(true, current), 2)
}

func renderHelp(bindings []key.Binding, rows int) string {
	var (
		// 1 pair of columns per binding: 1 for keys, 1 for descriptions
		// the number of pairs of columns is determined by the number of
		// bindings and the number of rows.
		cols      = make([]string, 2*int(math.Ceil(float64(len(bindings))/float64(rows))))
		keyStyle  = Regular.Copy().Bold(true).Align(lipgloss.Right).Margin(0, 1, 0, 2)
		descStyle = Regular.Copy().Align(lipgloss.Left)
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
