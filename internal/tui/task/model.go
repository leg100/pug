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
		tasks:    mm.Tasks,
		plans:    mm.Plans,
		task:     task,
		Helpers:  mm.Helpers,
		showInfo: mm.showInfo,
		width:    width,
		height:   height,
		program:  mm.Program,
	}
	if task.MachineReadableUI {
		m.sub = newMachineModel(task, machineModelOptions{
			width:  width,
			height: height,
		})
	} else {
		m.sub = newHuman(task, humanOptions{
			disableAutoscroll: mm.disableAutoscroll,
			spinner:           mm.Spinner,
			width:             width,
			height:            height,
		})
	}

	m.common = &tui.ActionHandler{
		Helpers:     mm.Helpers,
		IDRetriever: &m,
	}

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

	tasks  *task.Service
	task   *task.Task
	plans  *plan.Service
	common *tui.ActionHandler

	program  string
	showInfo bool

	sub tea.Model

	width  int
	height int
}

func (m *Model) Init() tea.Cmd {
	return m.sub.Init()
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
	case toggleTaskInfoMsg:
		m.showInfo = !m.showInfo
		// adjust width of sub model to accomodate info
		m.sub.Update(tea.WindowSizeMsg{
			Width:  m.subWidth(),
			Height: m.height,
		})
	case resource.Event[*task.Task]:
		if msg.Payload.ID != m.task.ID {
			// Ignore event for different task.
			return nil
		}
		m.task = msg.Payload
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.sub, cmd = m.sub.Update(tea.WindowSizeMsg{
			Width:  m.subWidth(),
			Height: m.height,
		})
		cmds = append(cmds, cmd)
	}

	// Handle remaining messages in sub model.
	m.sub, cmd = m.sub.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m Model) subWidth() int {
	if m.showInfo {
		return m.width - infoWidth
	}
	return m.width
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
	if !m.showInfo {
		return m.sub.View()
	}
	// Build info side pane
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
	return lipgloss.JoinHorizontal(lipgloss.Left, container, m.sub.View())
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
