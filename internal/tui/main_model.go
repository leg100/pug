package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
	"github.com/leg100/pug/internal/workspace"
)

type mainModel struct {
	*navigator

	width  int
	height int

	showHelp bool

	dump *os.File
}

type Options struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
	TaskService      *task.Service

	Logger    *logging.Logger
	Workdir   string
	FirstPage int
	MaxTasks  int
}

func New(opts Options) (mainModel, error) {
	messages, err := os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return mainModel{}, err
	}
	makers := map[modelKind]maker{
		ModuleListKind: &moduleListModelMaker{
			svc:        opts.ModuleService,
			workspaces: opts.WorkspaceService,
			workdir:    opts.Workdir,
		},
		WorkspaceListKind: &workspaceListModelMaker{
			svc:     opts.WorkspaceService,
			modules: opts.ModuleService,
		},
		RunListKind: &runListModelMaker{
			svc:   opts.RunService,
			tasks: opts.TaskService,
		},
		TaskListKind: &taskListModelMaker{
			svc:      opts.TaskService,
			maxTasks: opts.MaxTasks,
		},
		TaskKind: &taskModelMaker{
			svc: opts.TaskService,
		},
		LogsKind: &logsModelMaker{
			logger: opts.Logger,
		},
	}
	navigator, err := newNavigator(modelKind(opts.FirstPage), makers)
	if err != nil {
		return mainModel{}, err
	}
	m := mainModel{
		navigator: navigator,
		dump:      messages,
	}
	return m, nil
}

func (m mainModel) Init() tea.Cmd {
	return m.currentModel().Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	fmt.Fprintf(m.dump, "%#v\n", msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		cmds = append(cmds, m.resizeCmd)
	case resource.Event[any], common.ViewSizeMsg:
		// Send resource and view resize events to all cached models
		cmds = append(cmds, m.cache.updateAll(msg)...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.Keys.Quit):
			// ctrl-c quits the app
			return m, tea.Quit
		case key.Matches(msg, common.Keys.Escape):
			// <esc> whilst in help turns off help
			if m.showHelp {
				m.showHelp = false
			} else {
				m.goBack()
			}
		case key.Matches(msg, common.Keys.Help):
			// '?' toggles help
			m.showHelp = !m.showHelp
		case key.Matches(msg, common.Keys.Logs):
			// 'l' shows logs
			return m, navigate(page{kind: LogsKind})
		case key.Matches(msg, common.Keys.Modules):
			// 'm' lists all modules
			return m, navigate(page{kind: ModuleListKind})
		case key.Matches(msg, common.Keys.Workspaces):
			// 'W' lists all workspaces
			return m, navigate(page{kind: WorkspaceListKind})
		case key.Matches(msg, common.Keys.Runs):
			// 'R' lists all runs
			return m, navigate(page{kind: RunListKind})
		case key.Matches(msg, common.Keys.Tasks):
			// 'T' lists all tasks
			return m, navigate(page{kind: TaskListKind})
		}
	case navigationMsg:
		created, err := m.setCurrent(msg)
		if err != nil {
			return m, common.NewErrorCmd(err, "setting current page")
		}
		if created {
			cmds = append(cmds, m.currentModel().Init(), m.resizeCmd)
		}
	}
	// Send messages to current model
	cmd := m.updateCurrent(msg)
	return m, tea.Batch(append(cmds, cmd)...)
}

func (m mainModel) View() string {
	var (
		title   string
		content string
	)

	if m.showHelp {
		title = "help"
		content = renderHelp(m.currentModel().HelpBindings(), max(1, m.height-2))
	} else {
		title = m.currentModel().Title()
		content = m.currentModel().View()
	}

	top := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(common.DarkGrey)).
		Foreground(common.White).
		Padding(0, 1).
		Render(title)
	topWidth := lipgloss.Width(top)

	rows := []string{
		m.header(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			"─",
			top,
			strings.Repeat("─", max(0, m.width-topWidth)),
		),
		common.Regular.Copy().
			Height(m.viewHeight()).
			Width(m.viewWidth()).
			Render(content),
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
		Render(RenderShort(m.currentModel().HelpBindings()))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		help,
		renderedLogo,
	)
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

func (m mainModel) resizeCmd() tea.Msg {
	return common.ViewSizeMsg{Width: m.viewWidth(), Height: m.viewHeight()}
}
