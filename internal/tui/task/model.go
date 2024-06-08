package task

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/reflow/wordwrap"
)

// MakerID uniquely identifies a task model maker
type MakerID int

const (
	TaskMakerID MakerID = iota
	RunTabMakerID
	TaskListPreviewMakerID
	TaskGroupPreviewMakerID
)

type Maker struct {
	RunService  tui.RunService
	TaskService tui.TaskService
	Spinner     *spinner.Model
	MakerID     MakerID
	Helpers     *tui.Helpers
	Logger      *logging.Logger

	autoscroll bool
}

func (mm *Maker) Make(res resource.Resource, width, height int) (tea.Model, error) {
	task, ok := res.(*task.Task)
	if !ok {
		return model{}, errors.New("fatal: cannot make task model with non-task resource")
	}

	m := model{
		svc:     mm.TaskService,
		runs:    mm.RunService,
		task:    task,
		output:  task.NewReader(),
		spinner: mm.Spinner,
		makerID: mm.MakerID,
		// read upto 1kb at a time
		buf:        make([]byte, 1024),
		height:     height,
		helpers:    mm.Helpers,
		autoscroll: mm.autoscroll,
	}

	if rr := m.task.Run(); rr != nil {
		m.run = rr.(*run.Run)
	}

	m.viewport = viewport.New(0, 0)
	m.viewport.HighPerformanceRendering = false
	m.setWidth(width)
	m.setHeight(height)

	return m, nil
}

func (mm *Maker) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Autoscroll):
			mm.autoscroll = !mm.autoscroll

			var informUser tea.Cmd
			if mm.autoscroll {
				informUser = tui.ReportInfo("Enabled autoscroll")
			} else {
				informUser = tui.ReportInfo("Disabled autoscroll")
			}

			// Send out message to all cached task models to toggle autoscroll
			toggle := tui.CmdHandler(toggleAutoscrollMsg{})

			return tea.Batch(informUser, toggle)
		}
	}
	return nil
}

type model struct {
	svc  tui.TaskService
	task *task.Task
	run  *run.Run
	runs tui.RunService

	output     io.Reader
	buf        []byte
	content    string
	makerID    MakerID
	autoscroll bool

	viewport viewport.Model
	spinner  *spinner.Model

	height int

	showInfo bool

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
		// TODO: add keybinding to apply if task is a plan.
		case key.Matches(msg, localKeys.Info):
			// 'i' toggles showing task info
			m.showInfo = !m.showInfo
		case key.Matches(msg, keys.Common.Cancel):
			return m, m.helpers.CreateTasks("cancel", m.svc.Cancel, m.task.ID)
		case key.Matches(msg, keys.Common.Apply):
			if m.run != nil {
				// Only trigger an apply if run is in the planned state
				if m.run.Status != run.Planned {
					return m, nil
				}
				return m, tui.YesNoPrompt(
					"Apply run?",
					m.helpers.CreateTasks("apply", m.runs.ApplyPlan, m.run.ID),
				)
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
		m.content = wordwrap.String(m.content, m.viewport.Width)
		m.viewport.SetContent(m.content)
		if m.autoscroll {
			m.viewport.GotoBottom()
		}
		if msg.eof {
			if m.task.JSON {
				// Prettify JSON output from task. This can only be done once
				// the task has finished and has produced complete and
				// syntactically valid json object(s).
				//
				// TODO: avoid casting to string and back, thereby avoiding
				// unnecessary allocations.
				if b, err := prettyjson.Format([]byte(m.content)); err != nil {
					cmds = append(cmds, tui.ReportError(err, "pretty printing task json output"))
				} else {
					m.content = string(b)
					m.viewport.SetContent(string(b))
					if m.autoscroll {
						m.viewport.GotoBottom()
					}
				}
			}
			if m.content == "" {
				m.content = "Task finished without output"
				m.viewport.SetContent(m.content)
			}
		} else {
			cmds = append(cmds, m.getOutput)
		}
	case resource.Event[*task.Task]:
		if msg.Payload.ID != m.task.ID {
			// Ignore event for different task.
			return m, nil
		}
		m.task = msg.Payload
	case tea.WindowSizeMsg:
		m.setWidth(msg.Width)
		m.setHeight(msg.Height)
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

const (
	// scrollPercentWidth is the width of the scroll percentage section to the
	// right of the viewport
	scrollPercentWidth = 10
	// bordersWidth is the total width of the borders to the left and
	// right of the content
	bordersWidth = 2
	// bordersHeight is the total height of the borders to the top and
	// bottom of the content
	bordersHeight = 2
)

var borderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())

func (m *model) setWidth(width int) {
	viewportWidth := width - scrollPercentWidth
	if m.hasBorders() {
		viewportWidth -= bordersWidth
	}
	m.viewport.Width = max(0, viewportWidth)
}

func (m *model) setHeight(height int) {
	if m.hasBorders() {
		height -= bordersHeight
	}
	m.viewport.Height = height
	m.height = height
}

func (m *model) hasBorders() bool {
	return m.makerID == TaskMakerID
}

// View renders the viewport
func (m model) View() string {
	if m.showInfo {
		return strings.Join(m.task.Env, " ")
	}

	viewport := tui.Regular.Copy().
		MaxWidth(m.viewport.Width).
		Render(m.viewport.View())

	// scroll percent container occupies a fixed width section to the right of
	// the viewport.
	scrollPercent := tui.Regular.Copy().
		Background(tui.ScrollPercentageBackground).
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	scrollPercentContainer := tui.Regular.Copy().
		Margin(0, 1).
		Height(m.height).
		Width(scrollPercentWidth - 2).
		AlignVertical(lipgloss.Bottom).
		Render(scrollPercent)

	content := lipgloss.JoinHorizontal(lipgloss.Left,
		viewport,
		scrollPercentContainer,
	)

	if m.hasBorders() {
		return borderStyle.Render(content)
	}
	return content
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
