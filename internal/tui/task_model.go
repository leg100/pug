package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
)

type taskModelMaker struct {
	svc *task.Service
}

func (m *taskModelMaker) makeModel(taskResource resource.Resource) (common.Model, error) {
	task, err := m.svc.Get(taskResource.ID)
	if err != nil {
		return taskModel{}, err
	}
	return taskModel{
		svc:      m.svc,
		task:     task,
		output:   task.NewReader(),
		viewport: viewport.New(0, 0),
	}, nil
}

type taskOutputMsg string

type taskModel struct {
	svc    *task.Service
	task   *task.Task
	output io.Reader

	content  string
	viewport viewport.Model

	width  int
	height int
}

func NewTaskModel(svc *task.Service, taskID resource.ID, w, h int) (taskModel, error) {
	task, err := svc.Get(taskID)
	if err != nil {
		return taskModel{}, err
	}
	return taskModel{
		svc:      svc,
		task:     task,
		output:   task.NewReader(),
		viewport: viewport.New(w, h),
	}, nil
}

func (m taskModel) Init() tea.Cmd {
	return m.getOutput
}

func (m taskModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Cancel):
			return m, m.cancel
			// TODO: retry
		}
	case taskOutputMsg:
		m.content += string(msg)
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		if !m.task.IsFinished() {
			cmds = append(cmds, m.getOutput)
		}
	case resource.Event[*task.Task]:
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

func (m taskModel) Title() string {
	return "title goes here"
}

// View renders the viewport and the footer.
func (m taskModel) View() string {
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

func (m taskModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

func (m taskModel) getOutput() tea.Msg {
	out, err := io.ReadAll(m.output)
	if err != nil {
		return taskOutputMsg(err.Error())
	}
	return taskOutputMsg(string(out))
}

func (m taskModel) cancel() tea.Msg {
	if _, err := m.svc.Cancel(m.task.ID); err != nil {
		return common.NewErrorMsg(err, "canceling task")
	}
	return nil
}
