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
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/reflow/wordwrap"
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

func (mm *Maker) Make(id resource.ID, width, height int) (tea.Model, error) {
	return mm.make(id, width, height, true)
}

func (mm *Maker) make(id resource.ID, width, height int, border bool) (tea.Model, error) {
	task, err := mm.Tasks.Get(id)
	if err != nil {
		return model{}, err
	}

	m := model{
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
		border:   border,
		width:    width,
		program:  mm.Program,
	}
	m.setHeight(height)

	m.viewport = tui.NewViewport(tui.ViewportOptions{
		JSON:       m.task.JSON,
		Autoscroll: !mm.disableAutoscroll,
		Width:      m.viewportWidth(),
		Height:     m.height,
		Spinner:    m.spinner,
	})

	return m, nil
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

type model struct {
	*tui.Helpers

	id uuid.UUID

	tasks *task.Service
	task  *task.Task
	plans *plan.Service

	output <-chan []byte
	buf    []byte

	showInfo bool
	border   bool
	program  string

	viewport tui.Viewport
	spinner  *spinner.Model

	height int
	width  int
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
			return m, cancel(m.tasks, m.task.ID)
		case key.Matches(msg, keys.Common.Apply):
			spec, err := m.plans.ApplyPlan(m.task.ID)
			if err != nil {
				return m, tui.ReportError(err)
			}
			return m, tui.YesNoPrompt(
				"Apply plan?",
				m.CreateTasksWithSpecs(spec),
			)
		case key.Matches(msg, keys.Common.State):
			if ws := m.TaskWorkspaceOrCurrentWorkspace(m.task); ws != nil {
				return m, tui.NavigateTo(tui.ResourceListKind, tui.WithParent(ws.GetID()))
			} else {
				return m, tui.ReportError(errors.New("task not associated with a workspace"))
			}
		case key.Matches(msg, keys.Common.Retry):
			return m, tui.YesNoPrompt(
				"Retry task?",
				m.CreateTasksWithSpecs(m.task.Spec),
			)
		}
	case toggleAutoscrollMsg:
		m.viewport.Autoscroll = !m.viewport.Autoscroll
	case toggleTaskInfoMsg:
		m.showInfo = !m.showInfo
		// adjust width of viewport to accomodate info
		m.viewport.SetDimensions(m.viewportWidth(), m.height)
	case outputMsg:
		// Ensure output is for this model
		if msg.modelID != m.id {
			return m, nil
		}
		if err := m.viewport.AppendContent(msg.output, msg.eof); err != nil {
			return m, tui.ReportError(err)
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
		m.setHeight(msg.Height)
		m.viewport.SetDimensions(m.viewportWidth(), m.height)
		return m, nil
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) viewportWidth() int {
	if m.border {
		m.width -= 2
	}
	if m.showInfo {
		m.width -= infoWidth
	}
	return max(0, m.width)
}

func (m *model) setHeight(height int) {
	if m.border {
		height -= 2
	}
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
			fmt.Sprintf("Autoscroll: %s", boolToOnOff(m.viewport.Autoscroll)),
			"",
			fmt.Sprintf("Dependencies: %v", m.task.DependsOn),
		)

		// Word wrap task info to ensure it wraps "cleanly".
		wrapper := wordwrap.NewWriter(infoContentWidth)
		// Wrap on spaces and path separator
		wrapper.Breakpoints = []rune{' ', filepath.Separator}
		wrapper.Write([]byte(content))
		wrapped := wrapper.String()

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
	content := lipgloss.JoinHorizontal(lipgloss.Left, components...)
	if m.border {
		return tui.Border.Render(content)
	}
	return content
}

func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func (m model) Title() string {
	return m.Breadcrumbs("Task", m.task)
}

func (m model) Status() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.TaskSummary(m.task, false),
		m.TaskStatus(m.task, true),
	)
}

func (m model) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
		keys.Common.State,
		keys.Common.Retry,
		localKeys.ToggleInfo,
	}
	if moduleID := m.task.ModuleID; moduleID != nil {
		bindings = append(bindings, keys.Common.Module)
	}
	if workspaceID := m.task.WorkspaceID; workspaceID != nil {
		bindings = append(bindings, keys.Common.Workspace)
	}
	if m.task.Identifier == plan.ApplyTask {
		bindings = append(bindings, keys.Common.Apply)
	}
	return bindings
}

func (m model) getOutput() tea.Msg {
	msg := outputMsg{modelID: m.id}

	b, ok := <-m.output
	if ok {
		msg.output = b
	} else {
		msg.eof = true
	}
	return msg
}

type outputMsg struct {
	modelID uuid.UUID
	output  []byte
	eof     bool
}
