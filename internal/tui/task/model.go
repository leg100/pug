package task

import (
	"errors"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
)

// MakerID uniquely identifies a task model maker
type MakerID int

const (
	TaskMakerID MakerID = iota
	TaskListMakerID
	TaskGroupMakerID
)

type Maker struct {
	RunService  tui.RunService
	TaskService tui.TaskService
	Spinner     *spinner.Model
	Helpers     *tui.Helpers
	Logger      *logging.Logger

	disableAutoscroll bool
}

func (mm *Maker) Make(res resource.Resource, width, height int) (tea.Model, error) {
	return mm.make(res, width, height, TaskMakerID, tui.DefaultYPosition)
}

func (mm *Maker) make(res resource.Resource, width, height int, makerID MakerID, yPos int) (tea.Model, error) {
	task, ok := res.(*task.Task)
	if !ok {
		return model{}, errors.New("fatal: cannot make task model with a non-task resource")
	}

	m := model{
		Viewport: tui.NewViewport(tui.ViewportOptions{
			Width:     width,
			Height:    height,
			YPosition: yPos,
			JSON:      task.JSON,
		}),
		svc:     mm.TaskService,
		runs:    mm.RunService,
		task:    task,
		output:  task.NewReader(),
		spinner: mm.Spinner,
		makerID: makerID,
		// read upto 1kb at a time
		buf:        make([]byte, 1024),
		helpers:    mm.Helpers,
		autoscroll: !mm.disableAutoscroll,
	}

	if rr := m.task.Run(); rr != nil {
		m.run = rr.(*run.Run)
	}

	return m, nil
}

func (mm *Maker) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Autoscroll):
			mm.disableAutoscroll = !mm.disableAutoscroll

			// Inform user, and send out message to all cached task models to
			// toggle autoscroll.
			return tea.Batch(
				tui.CmdHandler(toggleAutoscrollMsg{}),
				tui.ReportInfo("Toggled autoscroll %s", tui.BoolToOnOff(!mm.disableAutoscroll)),
			)
		}
	}
	return nil
}

type model struct {
	tui.Viewport

	svc  tui.TaskService
	task *task.Task
	run  *run.Run
	runs tui.RunService

	output   io.Reader
	buf      []byte
	content  string
	makerID  MakerID
	finished bool

	autoscroll bool

	spinner *spinner.Model

	helpers *tui.Helpers
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.Viewport.Init()}
	if !m.finished {
		cmds = append(cmds, m.getOutput)
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Cancel):
			return m, m.helpers.CreateTasks("cancel", m.svc.Cancel, m.task.ID)
		case key.Matches(msg, keys.Common.Apply):
			if m.run != nil {
				// Only trigger an apply if run is in the planned state
				if m.run.Status != run.Planned {
					return m, nil
				}
				return m, tui.YesNoPrompt(
					"Apply plan?",
					m.helpers.CreateTasks("apply", m.runs.ApplyPlan, m.run.ID),
				)
			}
		case key.Matches(msg, keys.Common.State):
			if ws := m.helpers.TaskWorkspace(m.task); ws != nil {
				return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws))
			}
		}
	case toggleAutoscrollMsg:
		m.autoscroll = !m.autoscroll
	case outputMsg:
		// Ensure output is for this task
		if msg.taskID != m.task.ID {
			return m, nil
		}
		// Ensure output is for a task model made by the expected maker (avoids
		// duplicate output where there are multiple models for the same task).
		if msg.makerID != m.makerID {
			return m, nil
		}
		m.content += msg.output
		if msg.eof {
			if m.content == "" {
				m.content = "Task finished without output"
			}
			m.finished = true
		} else {
			cmds = append(cmds, m.getOutput)
		}
		if msg.output != "" {
			if err := m.SetContent(m.content, m.finished); err != nil {
				cmds = append(cmds, tui.ReportError(err, ""))
			}
			if m.autoscroll {
				m.GotoBottom()
			}
		}
	case resource.Event[*task.Task]:
		if msg.Payload.ID != m.task.ID {
			// Ignore event for different task.
			return m, nil
		}
		m.task = msg.Payload
	}

	// Handle keyboard and mouse events in the viewport
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return m.helpers.Breadcrumbs("Task", m.task)
}

func (m model) Status() string {
	if m.run != nil {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			m.helpers.TaskStatus(m.task, true),
			" | ",
			m.helpers.LatestRunReport(m.run),
			" ",
			m.helpers.RunStatus(m.run, true),
		)
	}
	return m.helpers.TaskStatus(m.task, true)
}

func (m model) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
	}
	if mod := m.task.Module(); mod != nil {
		bindings = append(bindings, keys.Common.Module)
	}
	if ws := m.task.Workspace(); ws != nil {
		bindings = append(bindings, keys.Common.Workspace)
	}
	if m.run != nil {
		bindings = append(bindings, keys.Common.Run)
		if m.run.Status == run.Planned {
			bindings = append(bindings, keys.Common.Apply)
		}
	}
	return bindings
}

func (m model) getOutput() tea.Msg {
	msg := outputMsg{taskID: m.task.ID, makerID: m.makerID}

	n, err := m.output.Read(m.buf)
	if err == io.EOF {
		msg.eof = true
	} else if err != nil {
		return tui.NewErrorMsg(err, "reading task output")
	}
	msg.output = string(m.buf[:n])
	return msg
}

type outputMsg struct {
	makerID MakerID
	taskID  resource.ID
	output  string
	eof     bool
}
