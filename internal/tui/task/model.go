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
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/reflow/wordwrap"
)

type Maker struct {
	RunService  tui.RunService
	TaskService tui.TaskService
	Spinner     *spinner.Model

	// If IsRunTab is true then Maker makes task models that are a tab within
	// the run model.
	IsRunTab bool

	Helpers *tui.Helpers
}

func (mm *Maker) Make(res resource.Resource, width, height int) (tea.Model, error) {
	task, ok := res.(*task.Task)
	if !ok {
		return model{}, errors.New("fatal: cannot make task model with non-task resource")
	}

	m := model{
		svc:      mm.TaskService,
		runs:     mm.RunService,
		task:     task,
		output:   task.NewReader(),
		spinner:  mm.Spinner,
		isRunTab: mm.IsRunTab,
		// read upto 1kb at a time
		buf:     make([]byte, 1024),
		width:   width,
		height:  height,
		helpers: mm.Helpers,
	}

	if rr := m.task.Run(); rr != nil {
		m.run = rr.(*run.Run)
	}

	m.viewport = viewport.New(0, 0)
	m.viewport.HighPerformanceRendering = false
	m.setViewportDimensions(width, height)

	return m, nil
}

type model struct {
	svc  tui.TaskService
	task *task.Task
	run  *run.Run
	runs tui.RunService

	output   io.Reader
	buf      []byte
	content  string
	isRunTab bool

	viewport viewport.Model
	spinner  *spinner.Model

	width  int
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
			return m, CreateTasks("cancel", m.task, m.svc.Cancel, m.task.ID)
		case key.Matches(msg, keys.Common.Module):
			// 'm' takes the user to the task's module, but only if the task
			// belongs to a module.
			if mod := m.task.Module(); mod != nil {
				return m, tui.NavigateTo(tui.ModuleKind, tui.WithParent(mod))
			}
		case key.Matches(msg, keys.Common.Workspace):
			// 'w' takes the user to the task's workspace, but only if the task
			// belongs to a workspace.
			if ws := m.task.Workspace(); ws != nil {
				return m, tui.NavigateTo(tui.WorkspaceKind, tui.WithParent(ws))
			}
		case key.Matches(msg, keys.Common.Run):
			// 'r' takes the user to the task's run, but only if the task
			// belongs to a run.
			if run := m.task.Run(); run != nil {
				return m, tui.NavigateTo(tui.RunKind, tui.WithParent(run))
			}
		case key.Matches(msg, keys.Common.Apply):
			if m.run != nil {
				// Only trigger an apply if run is in the planned state
				if m.run.Status != run.Planned {
					return m, nil
				}
				return m, tui.YesNoPrompt(
					"Apply run?",
					CreateTasks("apply", m.run, m.runs.ApplyPlan, m.run.ID),
				)
			}
		}
	case outputMsg:
		if msg.taskID != m.task.ID {
			return m, nil
		}
		// isRunTab is true when this msg is for a task model that is a child of a
		// run model, i.e. a tab. Without this flag, output would be duplicated in
		// both the tab and on the generic task view.
		if msg.isRunTab != m.isRunTab {
			return m, nil
		}
		m.content += msg.output
		m.content = wordwrap.String(m.content, m.width)
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		if msg.eof {
			if m.task.JSON {
				// Prettify JSON output from task. This can only be done once
				// the task has finished and has produced complete and
				// syntactically valid json object(s).
				//
				// Note: terraform commands such as `state -json` can produce
				// json strings with embedded newlines, which is invalid json
				// and breaks the pretty printer. So we escape the newlines.
				//
				// TODO: avoid casting to string and back, thereby avoiding
				// unnecessary allocations.
				m.content = strings.ReplaceAll(m.content, "\n", "\\n")
				if b, err := prettyjson.Format([]byte(m.content)); err != nil {
					cmds = append(cmds, tui.ReportError(err, "pretty printing task json output"))
				} else {
					m.content = string(b)
					m.viewport.SetContent(string(b))
					m.viewport.GotoBottom()
				}
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
		m.width = msg.Width
		m.height = msg.Height
		m.setViewportDimensions(msg.Width, msg.Height)
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
	// viewportMarginsWidth is the total width of the margins to the left and
	// right of the viewport
	viewportMarginsWidth = 2
)

func (m *model) setViewportDimensions(width, height int) {
	// minusWidth is the width to subtract from that available to the viewport
	minusWidth := scrollPercentWidth - viewportMarginsWidth

	// width is the available to the viewport
	width = max(0, width-minusWidth)

	m.viewport.Width = width
	m.viewport.Height = height
}

// View renders the viewport
func (m model) View() string {
	if m.showInfo {
		return strings.Join(m.task.Env, " ")
	}

	viewport := tui.Regular.Copy().
		Margin(0, 1).
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

	return lipgloss.JoinHorizontal(lipgloss.Left,
		viewport,
		scrollPercentContainer,
	)
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
	var color lipgloss.Color

	switch m.task.State {
	case task.Pending:
		color = tui.Grey
	case task.Queued:
		color = tui.Orange
	case task.Running:
		color = tui.Blue
	case task.Exited:
		color = tui.GreenBlue
	case task.Errored:
		color = tui.Red
	}
	taskState := tui.Regular.Copy().Background(color).Foreground(tui.White).Padding(0, 1).Render(string(m.task.State))

	if m.run != nil {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			m.helpers.LatestRunReport(m.run),
			" ",
			m.helpers.RunStatus(m.run),
		)
	}
	return taskState
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
	msg := outputMsg{taskID: m.task.ID, isRunTab: m.isRunTab}

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
	isRunTab bool
	taskID   resource.ID
	output   string
	eof      bool
}
