package internal

import (
	"io"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type logsModel struct {
	f io.Reader

	content  string
	viewport viewport.Model
}

func newLogsModel(t *task, mod module, w, h int) logsModel {
	f, err := os.Open("pug.log")
	return logsModel{
		f:        f,
		viewport: viewport.New(w, h),
	}
}

func (m logsModel) Init() tea.Cmd {
	return m.getTaskUpdate
}

func (m logsModel) Update(msg tea.Msg) (model, tea.Cmd) {
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

func (m logsModel) View() string {
	return lipgloss.NewStyle().
		Padding(0).
		Render(m.viewport.View())
}

func (m logsModel) getTaskUpdate() tea.Msg {
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

func (m logsModel) bindings() []key.Binding {
	return []key.Binding{keys.Modules}
}
