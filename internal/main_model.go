package internal

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewSizeMsg struct {
	width, height int
}

type state uint

const (
	modulesState state = iota
	taskState
)

type model interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (model, tea.Cmd)
	View() string
	bindings() []key.Binding
}

type mainModel struct {
	current state
	modules model
	task    model

	width, height int

	runner *runner
}

func NewMainModel(runner *runner) (mainModel, error) {
	mm, err := newModulesModel(runner)
	if err != nil {
		return mainModel{}, err
	}
	return mainModel{
		current: modulesState,
		modules: mm,
		runner:  runner,
	}, nil
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, func() tea.Msg {
			return viewSizeMsg{m.viewWidth(), m.viewHeight()}
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Modules):
			m.current = modulesState
			return m, nil
		}
	case newTaskMsg:
		m.current = taskState
		m.task = newTaskModel(msg.task, msg.mod, m.viewWidth(), m.viewHeight())
		return m, m.task.Init()
	case taskFailedMsg:
		// TODO: update a status bar
		return m, tea.Quit
	}

	switch m.current {
	case modulesState:
		newModel, cmd := m.modules.Update(msg)
		m.modules = newModel
		cmds = append(cmds, cmd)
	case taskState:
		newModel, cmd := m.task.Update(msg)
		m.task = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

var logo = strings.Join([]string{
	"▄▄▄ ▄ ▄ ▄▄▄",
	"█▄█ █ █ █ ▄",
	"▀   ▀▀▀ ▀▀▀",
}, "\n")

func (m mainModel) header(bindings []key.Binding) string {
	logo := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(darkgreen).
		Render(logo)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		logo,
		renderHelp(bindings),
	) + "\n"
}

func renderHelp(bindings []key.Binding) string {
	bindings = append(
		[]key.Binding{keys.Quit, keys.Help},
		bindings...,
	)
	var (
		// a column of keys and a column of descriptions for each group of three
		// bindings
		cols      = make([]string, 2*int(math.Ceil(float64(len(bindings))/3.)))
		keyStyle  = regular.Bold(true).Align(lipgloss.Right).Margin(0, 1, 0, 2)
		descStyle = regular.Align(lipgloss.Left)
	)

	for i := 0; i < len(bindings); i += 3 {
		rows := min(3, len(bindings)-i)
		keys := make([]string, rows)
		descs := make([]string, rows)
		for j := 0; j < rows; j++ {
			keys[j] = bindings[i+j].Help().Key
			descs[j] = bindings[i+j].Help().Desc
		}
		colnum := (i / 3) * 2
		cols[colnum] = keyStyle.Render(strings.Join(keys, "\n"))
		cols[colnum+1] = descStyle.Render(strings.Join(descs, "\n"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// viewHeight retrieves the height available within the main view
func (m mainModel) viewHeight() int {
	return m.height - 5
}

// viewWidth retrieves the width available within the main view
func (m mainModel) viewWidth() int {
	return m.width - roundedBorders.Copy().GetVerticalBorderSize()
}

func (m mainModel) View() string {
	viewbox := roundedBorders.Copy().Height(m.viewHeight()).Width(m.viewWidth())

	switch m.current {
	case modulesState:
		return m.header(m.modules.bindings()) + viewbox.Render(m.modules.View())
	case taskState:
		return m.header(m.task.bindings()) + viewbox.Render(m.task.View())
	default:
		return ""
	}
}
