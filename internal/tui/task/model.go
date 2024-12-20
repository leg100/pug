package task

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
)

type Maker struct {
	Plans   *plan.Service
	Tasks   *task.Service
	Spinner *spinner.Model
	Helpers *tui.Helpers
	Logger  *logging.Logger
	Program string

	disableAutoscroll bool
	showInfo          bool
}

func (mm *Maker) Make(id resource.ID, width, height int) (tui.ChildModel, error) {
	return mm.make(id, width, height, true)
}

func (mm *Maker) make(id resource.ID, width, height int, border bool) (tui.ChildModel, error) {
	task, err := mm.Tasks.Get(id)
	if err != nil {
		return nil, err
	}

	m := Model{
		id:      uuid.New(),
		tasks:   mm.Tasks,
		plans:   mm.Plans,
		task:    task,
		output:  task.NewStreamer(),
		spinner: mm.Spinner,
		// read upto 1kb at a time
		buf:      make([]byte, 1024),
		Helpers:  mm.Helpers,
		showInfo: mm.showInfo,
		width:    width,
		program:  mm.Program,
		// Disable autoscroll if either task is finished or user has disabled it
		disableAutoscroll: task.State.IsFinal() || mm.disableAutoscroll,
	}
	m.setHeight(height)

	m.viewport = tui.NewViewport(tui.ViewportOptions{
		JSON:    m.task.JSON,
		Width:   m.viewportWidth(),
		Height:  m.height,
		Spinner: m.spinner,
	})

	return &m, nil
}

func (mm *Maker) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Autoscroll):
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

type Model struct {
	*tui.Helpers

	id uuid.UUID

	tasks *task.Service
	task  *task.Task
	plans *plan.Service

	output <-chan []byte
	buf    []byte

	program           string
	disableAutoscroll bool
	showInfo          bool

	viewport tui.Viewport
	spinner  *spinner.Model

	height int
	width  int
}

func (m *Model) Init() tea.Cmd {
	return m.getOutput
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Cancel):
			return cancel(m.tasks, m.task.ID)
		case key.Matches(msg, keys.Common.Apply):
			spec, err := m.plans.ApplyPlan(m.task.ID)
			if err != nil {
				return tui.ReportError(err)
			}
			return tui.YesNoPrompt(
				"Apply plan?",
				m.CreateTasksWithSpecs(spec),
			)
		case key.Matches(msg, keys.Common.State):
			if ws := m.TaskWorkspaceOrCurrentWorkspace(m.task); ws != nil {
				return tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws.GetID()))
			} else {
				return tui.ReportError(errors.New("task not associated with a workspace"))
			}
		case key.Matches(msg, keys.Common.Retry):
			return tui.YesNoPrompt(
				"Retry task?",
				m.CreateTasksWithSpecs(m.task.Spec),
			)
		case key.Matches(msg, tui.Keys.SwitchPane):
			return tui.CmdHandler(tui.FocusExplorerMsg{})
		}
	case toggleAutoscrollMsg:
		m.disableAutoscroll = !m.disableAutoscroll
	case toggleTaskInfoMsg:
		m.showInfo = !m.showInfo
		// adjust width of viewport to accomodate info
		m.viewport.SetDimensions(m.viewportWidth(), m.height)
	case tui.OutputMsg:
		// Ensure output is for this model
		if msg.ModelID != m.id {
			return nil
		}
		err := m.viewport.AppendContent(msg.Output, msg.EOF, !m.disableAutoscroll)
		if err != nil {
			return tui.ReportError(err)
		}
		if !msg.EOF {
			cmds = append(cmds, m.getOutput)
		}
	case resource.Event[*task.Task]:
		if msg.Payload.ID != m.task.ID {
			// Ignore event for different task.
			return nil
		}
		m.task = msg.Payload
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.setHeight(msg.Height)
		m.viewport.SetDimensions(m.viewportWidth(), m.height)
		return nil
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m Model) viewportWidth() int {
	if m.showInfo {
		m.width -= infoWidth
	}
	return max(0, m.width)
}

func (m *Model) setHeight(height int) {
	m.height = height
}

const (
	// infoWidth is the width of the optional task info sidebar to the left of the
	// viewport.
	infoWidth = 40
	// infoContentWidth is the width available to the content inside the task
	// info sidebar, after subtracting 1 to accomodate its border to the right
	infoContentWidth = infoWidth - 1
)

// View renders the viewport
func (m *Model) View() string {
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
			tui.Bold.Render("Program"),
			m.task.Program,
			"",
			tui.Bold.Render("Arguments"),
			args,
			"",
			tui.Bold.Render("Path"),
			m.task.Path,
			"",
			tui.Bold.Render("Environment variables"),
			envs,
			"",
			fmt.Sprintf("Autoscroll: %s", boolToOnOff(!m.disableAutoscroll)),
			"",
			fmt.Sprintf("Dependencies: %v", m.task.DependsOn),
		)

		// Word wrap task info to ensure it wraps "cleanly".
		// Wrap on spaces and path separator
		wrapped := ansi.Wordwrap(content, infoContentWidth, " "+string(filepath.Separator))

		container := tui.Regular.
			Padding(0, 1).
			// Border to the right, dividing the info from the viewport
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(tui.LighterGrey).
			Height(m.height).
			// Crop content exceeding height
			MaxHeight(m.height).
			Width(infoContentWidth).
			Render(wrapped)
		components = append(components, container)
	}
	components = append(components, m.viewport.View())
	return lipgloss.JoinHorizontal(lipgloss.Left, components...)
}

func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func (m Model) BorderText() map[tui.BorderPosition]string {
	topRight := tui.Bold.Render(m.task.String())
	if path := m.TaskModulePathWithIcon(m.task); path != "" {
		topRight += " "
		topRight += path
	}
	if name := m.TaskWorkspaceNameWithIcon(m.task); name != "" {
		topRight += " "
		topRight += name
	}
	bottomLeft := m.TaskStatus(m.task, false)
	if summary := m.TaskSummary(m.task, false); summary != "" {
		bottomLeft += " "
		bottomLeft += summary
	}
	return map[tui.BorderPosition]string{
		tui.TopLeft:    topRight,
		tui.BottomLeft: bottomLeft,
	}
}

func (m Model) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
		keys.Common.State,
		keys.Common.Retry,
		localKeys.ToggleInfo,
	}
	if m.task.Identifier == plan.ApplyTask {
		bindings = append(bindings, keys.Common.Apply)
	}
	return bindings
}

func (m Model) getOutput() tea.Msg {
	msg := tui.OutputMsg{ModelID: m.id}

	b, ok := <-m.output
	if ok {
		msg.Output = b
	} else {
		msg.EOF = true
	}
	return msg
}
