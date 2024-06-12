package task

import (
	"errors"
	"fmt"
	"io"
	"strings"

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
	showInfo          bool
}

func (mm *Maker) Make(res resource.Resource, width, height int) (tea.Model, error) {
	return mm.makeWithID(res, width, height, TaskMakerID, true)
}

func (mm *Maker) makeWithID(res resource.Resource, width, height int, makerID MakerID, border bool) (tea.Model, error) {
	task, ok := res.(*task.Task)
	if !ok {
		return model{}, errors.New("fatal: cannot make task model with a non-task resource")
	}

	m := model{
		svc:     mm.TaskService,
		runs:    mm.RunService,
		task:    task,
		output:  task.NewReader(),
		spinner: mm.Spinner,
		makerID: makerID,
		// read upto 1kb at a time
		buf:      make([]byte, 1024),
		helpers:  mm.Helpers,
		showInfo: mm.showInfo,
		width:    width,
		height:   height,
	}

	if rr := m.task.Run(); rr != nil {
		m.run = rr.(*run.Run)
	}

	m.viewport = tui.NewViewport(tui.ViewportOptions{
		Width:      width,
		Height:     height,
		JSON:       true,
		Autoscroll: !mm.disableAutoscroll,
		Border:     border,
	})
	m.setWidth(width)
	m.setHeight(height)

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
				tui.ReportInfo("Toggled autoscroll %s", boolToOnOff(!mm.disableAutoscroll)),
			)
		case key.Matches(msg, localKeys.ToggleInfo):
			mm.showInfo = !mm.showInfo

			// Send out message to all cached task models to toggle task info
			return tui.CmdHandler(toggleTaskInfoMsg{})
		}
	}
	return nil
}

type model struct {
	svc  tui.TaskService
	task *task.Task
	run  *run.Run
	runs tui.RunService

	output  io.Reader
	buf     []byte
	makerID MakerID

	showInfo bool

	viewport tui.Viewport
	spinner  *spinner.Model

	height int
	width  int

	helpers *tui.Helpers
}

func (m model) Init() tea.Cmd {
	return m.getOutput
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
		m.viewport.Autoscroll = !m.viewport.Autoscroll
	case toggleTaskInfoMsg:
		m.showInfo = !m.showInfo
		// adjust width of viewport to accomodate info
		m.setWidth(m.width)
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
		if err := m.viewport.AppendContent(msg.output, msg.eof); err != nil {
			return m, tui.ReportError(err, "")
		}
		if !msg.eof {
			cmds = append(cmds, m.getOutput)
		}
	case resource.Event[*task.Task]:
		if msg.Payload.ID != m.task.ID {
			// Ignore event for different task.
			return m, nil
		}
		m.task = msg.Payload
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.viewport, _ = m.viewport.Update(tea.WindowSizeMsg{
			Width:  m.getViewportWidth(),
			Height: msg.Height,
		})
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) getViewportWidth() int {
	if m.showInfo {
		return m.width - infoWidth
	}
	return m.width
}

const (
	// bordersWidth is the total width of the borders to the left and
	// right of the content
	bordersWidth = 2
	// bordersHeight is the total height of the borders to the top and
	// bottom of the content
	bordersHeight = 2
	// infoWidth is the width of the optional task info sidebar to the left of the
	// viewport.
	infoWidth = 35
)

var borderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())

// View renders the viewport
func (m model) View() string {
	var components []string

	if m.showInfo {
		var (
			args = "-"
			envs = "-"
		)
		if len(m.task.Args) > 0 {
			args = strings.Join(m.task.Args, "\n")
		}
		if len(m.task.AdditionalEnv) > 0 {
			envs = strings.Join(m.task.AdditionalEnv, "\n")
		}

		// Show info to the left of the viewport.
		content := lipgloss.JoinVertical(lipgloss.Top,
			tui.Bold.Render("Task ID"),
			m.task.ID.String(),
			"",
			tui.Bold.Render("Command"),
			m.task.CommandString(),
			"",
			tui.Bold.Render("Arguments"),
			args,
			"",
			tui.Bold.Render("Environment variables"),
			envs,
			"",
			fmt.Sprintf("Autoscroll: %s", boolToOnOff(m.autoscroll)),
		)
		container := tui.Regular.Copy().
			Margin(0, 1).
			Height(m.height).
			// subtract 2 to account for margins, and 1 for the border to the
			// right
			Width(infoWidth-2-1).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(tui.LighterGrey).
			Render(content)
		components = append(components, container)
	}

	viewport := tui.Regular.Copy().
		Render(m.viewport.View())
	components = append(components, viewport)

	content := lipgloss.JoinHorizontal(lipgloss.Left, components...)

	return content
}

func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func (m model) TabStatus() string {
	switch m.task.State {
	case task.Running:
		return m.spinner.View()
	case task.Exited:
		return "✓"
	case task.Errored:
		return "✗"
	}
	return "+"
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
