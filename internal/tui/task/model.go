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
	Config  *Config
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
		buf:     make([]byte, 1024),
		Helpers: mm.Helpers,
		width:   width,
		program: mm.Program,
		config:  mm.Config,
		// Disable autoscroll if either task is finished or user has disabled it
		disableAutoscroll: task.State.IsFinal() || mm.Config.disableAutoscroll,
	}
	m.setHeight(height)

	m.viewport = tui.NewViewport(tui.ViewportOptions{
		JSON:    m.task.JSON,
		Width:   m.viewportWidth(),
		Height:  m.height,
		Spinner: m.spinner,
	})
	m.common = &tui.ActionHandler{
		Helpers:     mm.Helpers,
		IDRetriever: &m,
	}

	return &m, nil
}

type Model struct {
	*tui.Helpers

	id uuid.UUID

	tasks  *task.Service
	task   *task.Task
	plans  *plan.Service
	common *tui.ActionHandler

	output <-chan []byte
	buf    []byte

	program string

	viewport          tui.Viewport
	spinner           *spinner.Model
	config            *Config
	disableAutoscroll bool

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
		case key.Matches(msg, keys.Common.AutoApply):
			spec, err := m.plans.ApplyPlan(m.task.ID)
			if err != nil {
				return tui.ReportError(err)
			}
			return tui.YesNoPrompt(
				"Apply plan?",
				m.CreateTasksWithSpecs(spec),
			)
		case key.Matches(msg, keys.Common.Retry):
			return tui.YesNoPrompt(
				"Retry task?",
				m.CreateTasksWithSpecs(m.task.Spec),
			)
		default:
			cmd := m.common.Update(msg)
			cmds = append(cmds, cmd)
		}
	case toggleShowInfo:
		// adjust width of viewport to reflect presence/absence of task info
		// side pane.
		m.viewport.SetDimensions(m.viewportWidth(), m.height)
	case toggleAutoscrollMsg:
		m.disableAutoscroll = m.config.disableAutoscroll
	case outputMsg:
		// Ensure output is for this model
		if msg.modelID != m.id {
			return nil
		}
		err := m.viewport.AppendContent(msg.output, msg.eof, !m.disableAutoscroll)
		if err != nil {
			return tui.ReportError(err)
		}
		if !msg.eof {
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
	if m.config.showInfo {
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

	if m.config.showInfo {
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
			fmt.Sprintf("Autoscroll: %s", boolToOnOff(!m.config.disableAutoscroll)),
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
		tui.TopLeftBorder:    topRight,
		tui.BottomLeftBorder: bottomLeft,
	}
}

func (m Model) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		keys.Common.Cancel,
		keys.Common.State,
		keys.Common.Retry,
		localKeys.ToggleInfo,
	}
	if err := plan.IsApplyable(m.task); err == nil {
		bindings = append(bindings, localKeys.ApplyPlan)
	}
	bindings = append(bindings, m.common.HelpBindings()...)
	return bindings
}

func (m Model) getOutput() tea.Msg {
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

func (m Model) GetModuleIDs() ([]resource.ID, error) {
	if m.task.ModuleID == nil {
		return nil, errors.New("valid only on modules")
	}
	return []resource.ID{m.task.ModuleID}, nil
}

func (m Model) GetWorkspaceIDs() ([]resource.ID, error) {
	if m.task.WorkspaceID != nil {
		return []resource.ID{m.task.WorkspaceID}, nil
	} else if m.task.ModuleID == nil {
		return nil, errors.New("valid only on tasks associated with a module or a workspace")
	} else {
		// task has a module ID but no workspace ID, so find out if if
		// module has a current workspace, and if so, use that. Otherwise
		// return error
		mod, err := m.Modules.Get(m.task.ModuleID)
		if err != nil {
			return nil, err
		}
		if mod.CurrentWorkspaceID == nil {
			return nil, errors.New("valid only on tasks associated with a module with a current workspace, or a workspace")
		}
		return []resource.ID{mod.CurrentWorkspaceID}, nil
	}
}
