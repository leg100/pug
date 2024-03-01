package task

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/resource"
	taskpkg "github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
)

//func init() {
//	registerHelpBindings(func(short bool, current Page) []key.Binding {
//		if current != taskState {
//			return nil
//		}
//		if !short {
//			return keyMapToSlice(viewport.DefaultKeyMap())
//		}
//		return nil
//	})
//}

type (
	outputMsg string
)

type model struct {
	svc    *taskpkg.Service
	task   *taskpkg.Task
	output io.Reader

	content  string
	viewport viewport.Model

	width  int
	height int
}

func NewModel(svc *taskpkg.Service, task *taskpkg.Task, w, h int) model {
	return model{
		svc:      svc,
		task:     task,
		output:   task.NewReader(),
		viewport: viewport.New(w, h),
	}
}

func (m model) Init() tea.Cmd {
	return m.getOutput
}

func (m model) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Tasks, common.Keys.Escape):
			return m, common.Navigate(common.GlobalTaskListPage, uuid.UUID{})
		case key.Matches(msg, common.Keys.Cancel):
			return m, m.cancel
			// TODO: retry
		}
	case outputMsg:
		m.content += string(msg)
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		if !m.task.IsFinished() {
			cmds = append(cmds, m.getOutput)
		}
	case resource.Event[*taskpkg.Task]:
		if msg.Payload.ID == m.task.ID {
			m.task = msg.Payload
		}
	case common.ViewSizeMsg:
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

// Title is typically ignored in favour of the parent model's Title()
func (m model) Title() string {
	return ""
}

// View renders the viewport and the footer.
func (m model) View() string {
	status := lipgloss.NewStyle().
		Background(lipgloss.Color("#353533")).
		Foreground(common.White).
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
				Background(common.DarkGrey).
				Foreground(common.White).
				Width(m.width-common.Width(status)).
				Align(lipgloss.Right).
				Padding(0, 1).
				Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)),
		),
	)
}

func (m model) getOutput() tea.Msg {
	out, err := io.ReadAll(m.output)
	if err != nil {
		return outputMsg(err.Error())
	}
	return outputMsg(string(out))
}

func (m model) cancel() tea.Msg {
	if _, err := m.svc.Cancel(m.task.ID); err != nil {
		return common.NewErrorMsg(err, "canceling task")
	}
	return nil
}
