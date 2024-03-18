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
	"github.com/leg100/pug/internal/tui/table"
	"github.com/muesli/reflow/wordwrap"
)

type taskModelMaker struct {
	svc *task.Service
}

func (m *taskModelMaker) makeModel(tr resource.Resource) (Model, error) {
	return makeTaskModel(m.svc, tr, makeTaskModelOptions{})
}

type makeTaskModelOptions struct {
	// heightAdjustment adjusts the height of the rendered view by the given
	// number of rows.
	heightAdjustment int
	// isChild is true when the model is instantiated as a child tab of the run
	// model.
	isChild bool
}

func makeTaskModel(svc *task.Service, tr resource.Resource, opts makeTaskModelOptions) (taskModel, error) {
	task, err := svc.Get(tr.ID())
	if err != nil {
		return taskModel{}, err
	}

	return taskModel{
		svc:    svc,
		task:   task,
		output: task.NewReader(),
		// read upto 1kb at a time
		buf:              make([]byte, 1024),
		viewport:         viewport.New(0, 0),
		isChild:          opts.isChild,
		heightAdjustment: opts.heightAdjustment,
	}, nil
}

type taskOutputMsg struct {
	isChild bool
	taskID  resource.ID
	output  string
	eof     bool
}

type taskModel struct {
	svc  *task.Service
	task *task.Task

	output   io.Reader
	buf      []byte
	content  string
	viewport viewport.Model
	isChild  bool

	width            int
	height           int
	heightAdjustment int
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
		case key.Matches(msg, Keys.Apply):

			return m, taskCmd(m.svc.Cancel, m.task.ID())
			// TODO: retry
		}
	case taskOutputMsg:
		if msg.taskID != m.task.ID() {
			return m, nil
		}
		// isChild is true when this msg is for a task model that is a child of a
		// run model, i.e. a tab. Without this flag, output would be duplicated in
		// both the tab and on the generic task view.
		if msg.isChild != m.isChild {
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
	case common.ViewSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// subtract 2 to account for margins (1: left, 1: right)
		m.viewport.Width = msg.Width - 2
		// subtract 2 to account for task metadata rendered beneath viewport,
		// and make any necessary further adjustments according to whether
		// viewport is in a run tab or not.
		m.viewport.Height = msg.Height - 2 - m.heightAdjustment
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m taskModel) Title() string {
	return breadcrumbs("Task", *m.task.Parent)
}

// View renders the viewport
func (m taskModel) View() string {
	components := []string{
		Regular.Copy().
			Margin(0, 1).
			Render(m.viewport.View()),
	}
	// render task metadata only if task is not a child of a run model
	if !m.isChild {
		command := strings.Join(append(m.task.Command, m.task.Args...), " ")
		metadata := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("237")).
				Bold(true).
				Padding(0, 1).
				Render(command),
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("36")).
				Width(m.width-Width(command)-2).
				Align(lipgloss.Right).
				Padding(0, 1).
				Render(string(m.task.State)),
		)
		components = append(components,
			strings.Repeat("─", m.width),
			metadata,
		)
	}

	return lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				components...,
			),
		)
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

func (m taskModel) getOutput() tea.Msg {
	msg := taskOutputMsg{taskID: m.task.ID(), isChild: m.isChild}

	n, err := m.output.Read(m.buf)
	if err == io.EOF {
		msg.eof = true
	} else if err != nil {
		return newErrorMsg(err, "reading task output")
	}
	msg.output = string(m.buf[:n])
	return msg
}

// taskCmd returns a command that creates one or more tasks using the given IDs.
func taskCmd(fn task.Func, ids ...resource.ID) tea.Cmd {
	// Handle the case where a user has pressed a key on an empty table with
	// zero rows
	if len(ids) == 0 {
		return nil
	}

	// If items have been selected then clear the selection
	var deselectCmd tea.Cmd
	if len(ids) > 1 {
		deselectCmd = cmdHandler(table.DeselectMsg{})
	}

	cmd := func() tea.Msg {
		//var task *task.Task
		for _, id := range ids {
			var err error
			if _, err = fn(id); err != nil {
				return newErrorMsg(err, "creating task")
			}
		}
		//if len(ids) > 1 {
		//	// User has selected multiple rows, so send them to the task *list*
		//	// page
		//	//
		//	// TODO: pass in parameter specifying the parent resource for the
		//	// task listing, i.e. module, workspace, run, etc.
		//	return navigationMsg{
		//		target: page{kind: TaskListKind},
		//	}
		//} else {
		//	// User has highlighted a single row, so send them to the task page.
		//	return navigationMsg{
		//		target: page{kind: TaskKind, resource: task.Resource},
		//	}
		//}
		return nil
	}

	return tea.Batch(cmd, deselectCmd)
}
