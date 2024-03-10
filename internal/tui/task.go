package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/muesli/reflow/wordwrap"
)

type taskModelMaker struct {
	svc *task.Service
}

func (m *taskModelMaker) makeModel(taskResource resource.Resource) (Model, error) {
	task, err := m.svc.Get(taskResource.ID())
	if err != nil {
		return taskModel{}, err
	}

	spin := spinner.New()
	spin.Spinner = spinner.MiniDot
	spin.Style = lipgloss.NewStyle().
		Inherit(Breadcrumbs).
		Padding(0, 1)

	return taskModel{
		svc:    m.svc,
		task:   task,
		output: task.NewReader(),
		// read upto 1kb at a time
		buf:      make([]byte, 1024),
		viewport: viewport.New(0, 0),
		spinner:  spin,
	}, nil
}

type taskOutputMsg struct {
	taskID resource.ID
	output string
	eof    bool
}

type taskModel struct {
	svc    *task.Service
	task   *task.Task
	output io.Reader

	buf      []byte
	content  string
	viewport viewport.Model
	spinner  spinner.Model

	width  int
	height int
}

func (m taskModel) Init() tea.Cmd {
	return m.getOutput
}

func (m taskModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Cancel):
			return m, taskCmd(m.svc.Cancel, m.task.ID())
			// TODO: retry
		}
	case taskOutputMsg:
		if msg.taskID != m.task.ID() {
			return m, nil
		}
		m.content += msg.output
		m.content = wordwrap.String(m.content, m.width)
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		if !msg.eof {
			cmds = append(cmds, m.getOutput)
		}
	case resource.Event[*task.Task]:
		if msg.Payload.ID() != m.task.ID() {
			// Ignore event for different task.
			return m, nil
		}
		m.task = msg.Payload
		if m.task.State == task.Running {
			// Start the spinner once the task enters the running state.
			return m, m.spinner.Tick
		}
	case currentMsg:
		// Ensure spinner is spinning if user returns to this page.
		if m.task.State == task.Running {
			return m, m.spinner.Tick
		}
	case ViewSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// subtract 2 to account for margins (1: left, 1: right)
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height
	case spinner.TickMsg:
		// Keep spinning the spinner until the task stops.
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if !m.task.IsFinished() {
			return m, cmd
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m taskModel) Title() string {
	inherit := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(DarkGrey)).
		Foreground(White)

	// Module path
	components := []string{
		lipgloss.NewStyle().
			Inherit(inherit).
			Padding(0, 0, 0, 1).
			Render(m.task.Module().String()),
	}
	// Workspace
	if ws := m.task.Workspace(); ws != nil {
		components = append(components,
			lipgloss.NewStyle().
				Inherit(inherit).
				Render(ws.String()))
	}
	// Command
	components = append(components,
		lipgloss.NewStyle().
			Inherit(inherit).
			Align(lipgloss.Right).
			Render(strings.Join(m.task.Command, " ")))
	// Render breadcrumbs together
	breadcrumbs := strings.Join(components,
		lipgloss.NewStyle().
			Inherit(inherit).
			Render(" › "),
	)
	// spinner+status
	var renderedSpinner string
	if m.task.State == task.Running {
		renderedSpinner = m.spinner.View()
	}
	commandAndStatus := lipgloss.NewStyle().
		Width(m.width-max(0, Width(breadcrumbs))).
		Align(lipgloss.Right).
		Inherit(inherit).
		Padding(0, 1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				renderedSpinner,
				// status
				lipgloss.NewStyle().
					Inherit(inherit).
					Render(strings.ToUpper(string((m.task.State)))),
			),
		)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		breadcrumbs,
		commandAndStatus,
	)
}

// View renders the viewport
func (m taskModel) View() string {
	return lipgloss.NewStyle().
		Margin(0, 1).
		Render(m.viewport.View())
}

func (m taskModel) Pagination() string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#a8a7a5")).
		// off white
		Foreground(lipgloss.Color("#FAF9F6")).
		Padding(0, 1).
		Margin(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
}

func (m taskModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}

func (m *taskModel) getOutput() tea.Msg {
	msg := taskOutputMsg{taskID: m.task.ID()}

	n, err := m.output.Read(m.buf)
	if err == io.EOF {
		msg.eof = true
	} else if err != nil {
		return newErrorMsg(err, "reading task output")
	}
	msg.output = string(m.buf[:n])
	return msg
}

func taskCmd(fn func(resource.ID) (*task.Task, error), ids ...resource.ID) tea.Cmd {
	if len(ids) == 0 {
		return nil
	}

	// If items have been selected then clear the selection
	var deselectCmd tea.Cmd
	if len(ids) > 1 {
		deselectCmd = cmdHandler(deselectMsg{})
	}

	taskCmd := func() tea.Msg {
		var task *task.Task
		for _, id := range ids {
			var err error
			if task, err = fn(id); err != nil {
				return newErrorMsg(err, "creating task")
			}
		}
		if len(ids) > 1 {
			// User has selected multiple rows, so send them to the task *list*
			// page
			//
			// TODO: pass in parameter specifying the parent resource for the
			// task listing, i.e. module, workspace, run, etc.
			return navigationMsg{
				target: page{kind: TaskListKind},
			}
		} else {
			// User has highlighted a single row, so send them to the task page.
			return navigationMsg{
				target: page{kind: TaskKind, resource: task.Resource},
			}
		}
	}

	return tea.Batch(taskCmd, deselectCmd)
}
