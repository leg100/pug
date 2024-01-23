package internal

import (
	"strings"

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

type mainModel struct {
	current state
	modules tea.Model
	task    tea.Model

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
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "m": // go to modules view
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

func (m mainModel) header() string {
	column1 := lipgloss.NewStyle().Bold(true).Align(lipgloss.Right).Width(2).Margin(0, 1)
	column2 := lipgloss.NewStyle().Align(lipgloss.Left).Width(7).Margin(0, 0)
	var (
		keys  []string
		descs []string
	)
	for _, k := range defaultKeys {
		keys = append(keys, k.Help().Key)
		descs = append(descs, k.Help().Desc)
	}

	logo := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(darkgreen).
		Render(logo)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		logo,
		column1.Render(strings.Join(keys, "\n")),
		column2.Render(strings.Join(descs, "\n")),
	) + "\n"
}

// viewHeight retrieves the height available within the main view
func (m mainModel) viewHeight() int {
	return m.height - lipgloss.Height(m.header()) - 1
}

// viewWidth retrieves the width available within the main view
func (m mainModel) viewWidth() int {
	return m.width - roundedBorders.Copy().GetVerticalBorderSize()
}

func (m mainModel) View() string {
	var (
		borders = roundedBorders.Copy()
	)

	switch m.current {
	case modulesState:
		return m.header() + borders.Height(m.viewHeight()).Width(m.viewWidth()).Render(m.modules.View())
	case taskState:
		return m.header() + borders.Height(m.viewHeight()).Width(m.viewWidth()).Render(m.task.View())
	default:
		return ""
	}
}
