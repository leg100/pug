package top

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/common"
	"github.com/leg100/pug/internal/tui/logs"
	moduletui "github.com/leg100/pug/internal/tui/module"
	runtui "github.com/leg100/pug/internal/tui/run"
	tasktui "github.com/leg100/pug/internal/tui/task"
	workspacetui "github.com/leg100/pug/internal/tui/workspace"
	"github.com/leg100/pug/internal/workspace"
)

type model struct {
	*navigator

	width  int
	height int

	showHelp bool
	err      string

	tasks   *task.Service
	spinner *spinner.Model

	dump *os.File
}

type Options struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
	TaskService      *task.Service

	Logger    *logging.Logger
	Workdir   string
	FirstPage string
	MaxTasks  int
	Debug     bool
}

// New constructs the top-level TUI model.
func New(opts Options) (model, error) {
	firstKind, err := tui.FirstPageKind(opts.FirstPage)
	if err != nil {
		return model{}, err
	}

	var dump *os.File
	if opts.Debug {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return model{}, err
		}
	}

	spinner := spinner.New(spinner.WithSpinner(spinner.Globe))

	taskMaker := &tasktui.Maker{
		TaskService: opts.TaskService,
	}

	makers := map[tui.Kind]tui.Maker{
		tui.ModuleListKind: &moduletui.ListMaker{
			ModuleService: opts.ModuleService,
			RunService:    opts.RunService,
			Spinner:       &spinner,
			Workdir:       opts.Workdir,
		},
		tui.WorkspaceListKind: &workspacetui.ListMaker{
			WorkspaceService: opts.WorkspaceService,
			ModuleService:    opts.ModuleService,
			RunService:       opts.RunService,
		},
		tui.RunListKind: &runtui.ListMaker{
			RunService:  opts.RunService,
			TaskService: opts.TaskService,
		},
		tui.TaskListKind: &tasktui.ListMaker{
			TaskService: opts.TaskService,
			MaxTasks:    opts.MaxTasks,
		},
		tui.RunKind: &runtui.Maker{
			RunService:  opts.RunService,
			TaskService: opts.TaskService,
			Spinner:     &spinner,
		},
		tui.TaskKind: taskMaker,
		tui.LogsKind: &logs.Maker{
			Logger: opts.Logger,
		},
	}
	navigator, err := newNavigator(firstKind, makers)
	if err != nil {
		return model{}, err
	}
	m := model{
		navigator: navigator,
		spinner:   &spinner,
		tasks:     opts.TaskService,
		dump:      dump,
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return m.currentModel().Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}

	// Keep shared spinner spinning as long as there are tasks running.
	switch msg := msg.(type) {
	case resource.Event[*task.Task]:
		if m.tasks.Counter() > 0 {
			cmds = append(cmds, m.spinner.Tick)
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		*m.spinner, cmd = m.spinner.Update(msg)
		if m.tasks.Counter() > 0 {
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Inform navigator of new dimenisions for when it builds new models
		m.navigator.width = m.viewWidth()
		m.navigator.height = m.viewHeight()

		// Send out new message with adjusted dimensions
		return m, func() tea.Msg {
			return common.ViewSizeMsg{Width: m.viewWidth(), Height: m.viewHeight()}
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Quit):
			// ctrl-c quits the app
			return m, tea.Quit
		case key.Matches(msg, tui.Keys.Escape):
			// <esc> closes help or goes back to last page
			if m.showHelp {
				m.showHelp = false
			} else {
				m.goBack()
			}
		case key.Matches(msg, tui.Keys.Help):
			// '?' toggles help
			m.showHelp = !m.showHelp
		case key.Matches(msg, tui.Keys.Logs):
			// 'l' shows logs
			return m, tui.NavigateTo(tui.LogsKind, nil)
		case key.Matches(msg, tui.Keys.Modules):
			// 'm' lists all modules
			return m, tui.NavigateTo(tui.ModuleListKind, nil)
		case key.Matches(msg, tui.Keys.Workspaces):
			// 'W' lists all workspaces
			return m, tui.NavigateTo(tui.WorkspaceListKind, nil)
		case key.Matches(msg, tui.Keys.Runs):
			// 'R' lists all runs
			return m, tui.NavigateTo(tui.RunListKind, nil)
		case key.Matches(msg, tui.Keys.Tasks):
			// 'T' lists all tasks
			return m, tui.NavigateTo(tui.TaskListKind, nil)
		default:
			// Send other keys to current model.
			cmd := m.updateCurrent(msg)
			return m, cmd
		}
	case tui.NavigationMsg:
		created, err := m.setCurrent(tui.Page(msg))
		if err != nil {
			return m, tui.NewErrorCmd(err, "setting current page")
		}
		if created {
			return m, m.currentModel().Init()
		}
		return m, nil
	case tui.ErrorMsg:
		if msg.Error != nil {
			m.err = fmt.Sprintf("%s: %s", fmt.Sprintf(msg.Message, msg.Args...), msg.Error.Error())
		}
	default:
		// Send remaining msg types to all cached models
		cmds = append(cmds, m.cache.updateAll(msg)...)
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

var (
	logo = strings.Join([]string{
		"▄▄▄ ▄ ▄ ▄▄▄",
		"█▄█ █ █ █ ▄",
		"▀   ▀▀▀ ▀▀▀",
	}, "\n")
	renderedLogo = tui.Bold.
			Copy().
			Margin(0, 1).
			Foreground(tui.Pink).
			Render(logo)
	logoWidth            = lipgloss.Width(renderedLogo)
	headerHeight         = 3
	breadcrumbsHeight    = 1
	horizontalRuleHeight = 1
	messageFooterHeight  = 1
)

func (m model) View() string {
	var (
		content           string
		shortHelpBindings []key.Binding
		pagination        = tui.Regular.Padding(0, 1).Render(
			fmt.Sprintf("%d/%d tasks", m.tasks.Counter(), 32),
		)
	)

	if m.showHelp {
		content = lipgloss.NewStyle().
			Margin(1).
			Render(
				fullHelpView(
					m.currentModel().HelpBindings(),
					tui.KeyMapToSlice(tui.GeneralKeys),
					tui.KeyMapToSlice(viewport.DefaultKeyMap()),
				),
			)
		shortHelpBindings = []key.Binding{
			key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "close help"),
			),
		}
	} else {
		content = m.currentModel().View()
		shortHelpBindings = append(
			m.currentModel().HelpBindings(),
			tui.KeyMapToSlice(tui.GeneralKeys)...,
		)
	}

	// Center title within a horizontal rule
	title := m.currentModel().Title()
	titleRemainingWidth := m.width - tui.Width(title)
	titleRemainingWidthHalved := titleRemainingWidth / 2
	titleLeftRule := strings.Repeat("─", max(0, titleRemainingWidthHalved))
	titleLeftRuleAndTitle := fmt.Sprintf("%s %s ", titleLeftRule, title)
	titleRightRule := strings.Repeat("─", max(0, m.width-tui.Width(titleLeftRuleAndTitle)))
	renderedTitle := fmt.Sprintf("%s%s", titleLeftRuleAndTitle, titleRightRule)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		// header
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			// help
			lipgloss.NewStyle().
				Margin(0, 1).
				// -2 for vertical margins
				Width(m.width-logoWidth-2).
				Render(shortHelpView(shortHelpBindings, m.width-logoWidth-2)),
			// logo
			lipgloss.NewStyle().
				Render(renderedLogo),
		),
		// title
		lipgloss.NewStyle().
			// Prohibit overflowing title wrapping to another line.
			MaxHeight(1).
			Width(m.width).
			Render(renderedTitle),
		// content
		lipgloss.NewStyle().
			Height(m.viewHeight()).
			Render(content),
		// horizontal rule
		lipgloss.NewStyle().
			Render(strings.Repeat("─", m.width)),
		// footer
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			// error messages
			lipgloss.NewStyle().
				Width(m.width-tui.Width(pagination)).
				Padding(0, 1).
				Padding(0, 1).
				Foreground(tui.Red).
				Render(m.err),
			// pagination
			pagination,
		),
	)
}

// viewHeight retrieves the height available beneath the header and breadcrumbs,
// and the message footer.
func (m model) viewHeight() int {
	return m.height - headerHeight - breadcrumbsHeight - horizontalRuleHeight - messageFooterHeight
}

// viewWidth retrieves the width available within the main view
func (m model) viewWidth() int {
	return m.width
}
