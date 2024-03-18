package tui

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
	"github.com/leg100/pug/internal/tui/common"
	"github.com/leg100/pug/internal/workspace"
)

type mainModel struct {
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

func New(opts Options) (mainModel, error) {
	firstKind, err := firstPageKind(opts.FirstPage)
	if err != nil {
		return mainModel{}, err
	}

	var dump *os.File
	if opts.Debug {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return mainModel{}, err
		}
	}

	spinner := spinner.New(spinner.WithSpinner(spinner.Globe))

	taskModelMaker := &taskModelMaker{
		svc: opts.TaskService,
	}

	makers := map[modelKind]maker{
		ModuleListKind: &moduleListModelMaker{
			svc:        opts.ModuleService,
			workspaces: opts.WorkspaceService,
			runs:       opts.RunService,
			spinner:    &spinner,
			workdir:    opts.Workdir,
		},
		WorkspaceListKind: &workspaceListModelMaker{
			svc:     opts.WorkspaceService,
			modules: opts.ModuleService,
			runs:    opts.RunService,
		},
		RunListKind: &runListModelMaker{
			svc:   opts.RunService,
			tasks: opts.TaskService,
		},
		TaskListKind: &taskListModelMaker{
			svc:      opts.TaskService,
			maxTasks: opts.MaxTasks,
		},
		RunKind: &runModelMaker{
			svc:     opts.RunService,
			tasks:   opts.TaskService,
			spinner: &spinner,
		},
		TaskKind: taskModelMaker,
		LogsKind: &logsModelMaker{
			logger: opts.Logger,
		},
	}
	navigator, err := newNavigator(firstKind, makers)
	if err != nil {
		return mainModel{}, err
	}
	m := mainModel{
		navigator: navigator,
		spinner:   &spinner,
		tasks:     opts.TaskService,
		dump:      dump,
	}
	return m, nil
}

func (m mainModel) Init() tea.Cmd {
	return m.currentModel().Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}

	if m.showHelp {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			return m, m.resizeCmd
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keys.Escape, Keys.CloseHelp):
				// <esc>, '?' closes help
				m.showHelp = false
				return m, nil
			}
		}
	}

	// Keep tally of the number of running tasks.
	switch msg := msg.(type) {
	case resource.Event[*task.Task]:
		if m.tasks.Counter() > 0 {
			cmds = append(cmds, m.spinner.Tick)
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		*m.spinner, cmd = m.spinner.Update(msg)
		// Only continue spinning the spinner while there are tasks running
		if m.tasks.Counter() > 0 {
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.resizeCmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			// ctrl-c quits the app
			return m, tea.Quit
		case key.Matches(msg, Keys.Escape):
			m.goBack()
		case key.Matches(msg, Keys.Help):
			// '?' shows help
			m.showHelp = true
		case key.Matches(msg, Keys.Logs):
			// 'l' shows logs
			return m, navigate(page{kind: LogsKind})
		case key.Matches(msg, Keys.Modules):
			// 'm' lists all modules
			return m, navigate(page{kind: ModuleListKind})
		case key.Matches(msg, Keys.Workspaces):
			// 'W' lists all workspaces
			return m, navigate(page{kind: WorkspaceListKind})
		case key.Matches(msg, Keys.Runs):
			// 'R' lists all runs
			return m, navigate(page{kind: RunListKind})
		case key.Matches(msg, Keys.Tasks):
			// 'T' lists all tasks
			return m, navigate(page{kind: TaskListKind})
		default:
			// Send other keys to current model.
			cmd := m.updateCurrent(msg)
			return m, cmd
		}
	case navigationMsg:
		created, err := m.setCurrent(msg)
		if err != nil {
			return m, newErrorCmd(err, "setting current page")
		}
		if created {
			return m, tea.Batch(m.currentModel().Init(), m.resizeCmd)
		}
		return m, m.resizeCmd
	case errorMsg:
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
	renderedLogo = Bold.
			Copy().
			Margin(0, 1).
			Foreground(Pink).
			Render(logo)
	logoWidth            = lipgloss.Width(renderedLogo)
	headerHeight         = 3
	breadcrumbsHeight    = 1
	horizontalRuleHeight = 1
	messageFooterHeight  = 1
)

// breadcrumbs renders the breadcrumbs for a page, i.e. the ancestry of the
// page's resource.
func breadcrumbs(title string, parent resource.Resource) string {
	// format: <title>(<path>:<workspace>:<run>)
	var crumbs []string
	switch parent.Kind {
	case resource.Run:
		// if parented by a run, then include its ID
		runID := Regular.Copy().Foreground(lightGrey).Render(parent.Run().String())
		crumbs = append([]string{fmt.Sprintf("{%s}", runID)}, crumbs...)
		fallthrough
	case resource.Workspace:
		// if parented by a workspace, then include its name
		name := Regular.Copy().Foreground(Red).Render(parent.Workspace().String())
		crumbs = append([]string{fmt.Sprintf("[%s]", name)}, crumbs...)
		fallthrough
	case resource.Module:
		// if parented by a module, then include its path
		path := Regular.Copy().Foreground(Blue).Render(parent.Module().String())
		crumbs = append([]string{fmt.Sprintf("(%s)", path)}, crumbs...)
	case resource.Global:
		// if parented by global, then state it is global
		global := Regular.Copy().Foreground(Blue).Render("global")
		crumbs = append([]string{fmt.Sprintf("(%s)", global)}, crumbs...)
	}
	return fmt.Sprintf("%s%s", Bold.Render(title), strings.Join(crumbs, ""))
}

func (m mainModel) View() string {
	var (
		content           string
		shortHelpBindings []key.Binding
		pagination        = Regular.Padding(0, 1).Render(
			fmt.Sprintf("%d/%d tasks", m.tasks.Counter(), 32),
		)
	)

	if m.showHelp {
		content = lipgloss.NewStyle().
			Margin(1).
			Render(
				fullHelpView(
					m.currentModel().HelpBindings(),
					keyMapToSlice(generalKeys),
					keyMapToSlice(viewport.DefaultKeyMap()),
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
			keyMapToSlice(generalKeys)...,
		)
	}

	// Center title within a horizontal rule
	title := m.currentModel().Title()
	titleRemainingWidth := m.width - Width(title)
	titleRemainingWidthHalved := titleRemainingWidth / 2
	titleLeftRule := strings.Repeat("─", max(0, titleRemainingWidthHalved))
	titleLeftRuleAndTitle := fmt.Sprintf("%s %s ", titleLeftRule, title)
	titleRightRule := strings.Repeat("─", max(0, m.width-Width(titleLeftRuleAndTitle)))
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
				Width(m.width-Width(pagination)).
				Padding(0, 1).
				Padding(0, 1).
				Foreground(Red).
				Render(m.err),
			// pagination
			pagination,
		),
	)
}

// viewHeight retrieves the height available beneath the header and breadcrumbs,
// and the message footer.
func (m mainModel) viewHeight() int {
	return m.height - headerHeight - breadcrumbsHeight - horizontalRuleHeight - messageFooterHeight
}

// viewWidth retrieves the width available within the main view
func (m mainModel) viewWidth() int {
	return m.width
}

func (m mainModel) resizeCmd() tea.Msg {
	return common.ViewSizeMsg{Width: m.viewWidth(), Height: m.viewHeight()}
}
