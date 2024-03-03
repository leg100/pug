package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
	"github.com/leg100/pug/internal/workspace"
)

type mainModel struct {
	TaskService      *task.Service
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
	Workdir          string

	last            common.Model
	current         common.Model
	moduleModel     common.Model
	workspacesModel common.Model
	tasksModel      common.Model
	logsModel       common.Model

	width  int
	height int

	// status contains extraordinary info, e.g. errors, warnings
	status string
}

type Options struct {
	TaskService      *task.Service
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
	Workdir          string
}

func New(opts Options) (mainModel, error) {
	moduleModel, err := NewModuleListModel(opts.ModuleService, opts.Workdir)
	if err != nil {
		return mainModel{}, err
	}

	return mainModel{
		moduleModel: moduleModel,
		logsModel:   newLogsModel(),
		current:     moduleModel,
	}, nil
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.resizeCmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, common.Keys.Help):
			// open help, keeping reference to last state and passing in the
			// bindings for the last state.
			m.last = m.current
			m.current = newHelpModel(m.last.HelpBindings())
			return m, nil
		case key.Matches(msg, common.Keys.Logs):
			// open logs, keeping reference to last state
			m.last = m.current
			m.current = m.logsModel
			return m, nil
		case key.Matches(msg, common.Keys.Modules):
			m.current = m.moduleModel
			return m, nil
		case key.Matches(msg, common.Keys.Workspaces):
			m.current = m.workspacesModel
			return m, nil
		case key.Matches(msg, common.Keys.Tasks):
			m.current = m.tasksModel
			return m, nil
		default:
			// send key to current model
			updated, cmd := m.current.Update(msg)
			m.current = updated
			return m, cmd
		}
	case common.ReturnLastMsg:
		m.current = m.last
		return m, nil
	case common.NavigationMsg:
		var (
			to  common.Model
			err error
		)
		switch msg.To {
		case common.LogsPage:
			to = tuitask.NewTaskListModel(m.TaskService, msg.Resource)
		case common.HelpPage:
			to = tuitask.NewTaskListModel(m.TaskService, msg.Resource)
		case common.ModuleListPage:
			to, err = tuimodule.NewModuleListModel(m.ModuleService, m.Workdir)
		case common.RunListPage:
			to = tuirun.NewListModel(m.RunService, msg.Resource)
		case common.WorkspaceListPage:
			to = tuiworkspace.NewListModel(m.WorkspaceService, msg.Resource)
		case common.TaskListPage:
			to = tuitask.NewTaskListModel(m.TaskService, msg.Resource)
		case common.TaskPage:
			to, err = tuitask.NewTaskModel(m.TaskService, msg.Resource.ID, m.width, m.height)
		}
		if err != nil {
			return m, common.NewErrorCmd(err, "navigating pages")
		}
		m.current = to
		cmds = append(cmds, to.Init(), m.resizeCmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) resizeCmd() tea.Msg {
	return common.ViewSizeMsg{Width: m.viewWidth(), Height: m.viewHeight()}
}

// viewHeight retrieves the height available within the main view
func (m mainModel) viewHeight() int {
	// hardcode height adjustment for performance reasons:
	// heading: 3
	// hr: 1
	// title: 1
	// hr: 1
	return m.height - 4
}

// viewWidth retrieves the width available within the main view
func (m mainModel) viewWidth() int {
	return m.width
}

func (m mainModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(common.DarkGrey)).
		Foreground(common.White).
		Padding(0, 1).
		Render(m.current.Title())
	titleWidth := lipgloss.Width(title)

	rows := []string{
		m.header(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			"─",
			title,
			strings.Repeat("─", max(0, m.width-titleWidth)),
		),
		common.Regular.Copy().
			Height(m.viewHeight()).
			Width(m.viewWidth()).
			Render(m.current.View()),
	}
	return lipgloss.JoinVertical(lipgloss.Top, rows...)
}

var (
	logo = strings.Join([]string{
		"▄▄▄ ▄ ▄ ▄▄▄",
		"█▄█ █ █ █ ▄",
		"▀   ▀▀▀ ▀▀▀",
	}, "\n")
	renderedLogo = common.Bold.
			Copy().
			Padding(0, 1).
			Foreground(common.Pink).
			Render(logo)
	logoWidth = lipgloss.Width(renderedLogo)
)

func (m mainModel) header() string {
	help := lipgloss.NewStyle().
		Width(m.width - logoWidth).
		Render(RenderShort(m.current))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		help,
		renderedLogo,
	)
}
