package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	taskpkg "github.com/leg100/pug/internal/task"
)

func init() {
	registerHelpBindings(func(short bool, current State) []key.Binding {
		if current != taskState {
			return nil
		}
		if !short {
			return keyMapToSlice(viewport.DefaultKeyMap())
		}
		return nil
	})
}

const taskState State = "task"

type (
	taskNewMsg    *taskpkg.Task
	taskUpdateMsg string
	taskEventMsg  struct{}
	taskFailedMsg string
)

type task struct {
	task *taskpkg.Task
	mod  *module.Module

	content  string
	viewport viewport.Model

	width  int
	height int
}

func newTask(t *taskpkg.Task, mod *module.Module, w, h int) task {
	return task{
		task:     t,
		mod:      mod,
		viewport: viewport.New(w, h),
	}
}

func (m task) Init() tea.Cmd {
	return tea.Batch(m.getTaskUpdate, m.watchTaskEvents)
}

func (m task) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Tasks, Keys.Escape):
			return m, ChangeState(tasksState, WithModelOption(
				newTasks(m.mod),
			))
		case key.Matches(msg, Keys.Retry):
			return m, ChangeState(tasksState, WithModelOption(
				newTasks(m.mod),
			))
		}
	case taskUpdateMsg:
		m.content += string(msg)
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		if !m.task.IsFinished() {
			cmds = append(cmds, m.getTaskUpdate)
		}
	case taskEventMsg:
		if !m.task.IsFinished() {
			cmds = append(cmds, m.watchTaskEvents)
		}
	case taskFailedMsg:
		m.content = string(msg)
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
	case ViewSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// subtract 2 to account for margins
		m.viewport.Width = msg.Width - 2
		// subtract 1 to account for status bar
		m.viewport.Height = msg.Height - 1
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m task) Title() string {
	workspace := lipgloss.NewStyle().
		Background(lipgloss.Color(DarkGrey)).
		Foreground(White).
		Render("default")
	return fmt.Sprintf("task (%s) %s", m.mod.Path, workspace)
}

func (m task) View() string {
	status := lipgloss.NewStyle().
		Background(lipgloss.Color("#353533")).
		Foreground(White).
		Padding(0, 1).
		Render(strings.ToUpper(string((m.task.State))))

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().
			Margin(0, 1).
			Render(m.viewport.View()),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			status,
			lipgloss.NewStyle().
				Background(DarkGrey).
				Foreground(White).
				Width(m.width-width(status)).
				Align(lipgloss.Right).
				Padding(0, 1).
				Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)),
		),
	)
}

func (m task) getTaskUpdate() tea.Msg {
	out, err := io.ReadAll(m.task.Output)
	if err != nil {
		return taskUpdateMsg(err.Error())
	}
	return taskUpdateMsg(string(out))
}

func (m task) watchTaskEvents() tea.Msg {
	<-m.task.Events
	return taskEventMsg{}
}
