package internal

import (
	"io"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type newTaskMsg struct {
	mod  module
	task *task
}

type taskFailedMsg string

type taskUpdateMsg struct {
	content string
	ended   bool
}

type taskModel struct {
	task *task
	mod  module

	content  string
	viewport viewport.Model
}

func newTaskModel(t *task, mod module, w, h int) taskModel {
	return taskModel{
		task:     t,
		mod:      mod,
		viewport: viewport.New(w, h),
	}
}

func (m taskModel) Init() tea.Cmd {
	return m.getTaskUpdate
}

func (m taskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case taskUpdateMsg:
		m.content += msg.content
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		if !msg.ended {
			cmds = append(cmds, m.getTaskUpdate)
		}
	case viewSizeMsg:
		m.viewport.Width = msg.width
		m.viewport.Height = msg.height
	default:
		// Handle keyboard and mouse events in the viewport
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m taskModel) View() string {
	return lipgloss.NewStyle().
		Padding(0).
		Render(m.viewport.View())
}

func (m taskModel) getTaskUpdate() tea.Msg {
	p := make([]byte, 100)
	if _, err := m.task.out.Read(p); err == io.EOF {
		return taskUpdateMsg{
			ended: true,
		}
	} else if err != nil {
		return taskUpdateMsg{
			content: err.Error(),
		}
	}
	return taskUpdateMsg{
		content: string(p),
	}
}
