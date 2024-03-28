package top

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/version"
	"github.com/leg100/pug/internal/workspace"
)

type model struct {
	WorkspaceService *workspace.Service

	*navigator

	width  int
	height int

	showHelp bool

	// Either an error or an informational message is rendered in the footer.
	err  error
	info string

	tasks   *task.Service
	spinner *spinner.Model

	dump *os.File

	workdir string
}

type Options struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	StateService     *state.Service
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
	var dump *os.File
	if opts.Debug {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return model{}, err
		}
	}

	workdir, err := contractUserPath(opts.Workdir)
	if err != nil {
		return model{}, err
	}

	spinner := spinner.New(spinner.WithSpinner(spinner.Globe))

	navigator, err := newNavigator(opts, &spinner)
	if err != nil {
		return model{}, err
	}

	m := model{
		WorkspaceService: opts.WorkspaceService,
		navigator:        navigator,
		spinner:          &spinner,
		tasks:            opts.TaskService,
		dump:             dump,
		workdir:          workdir,
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.currentModel().Init(),
		tui.ReloadModules(m.WorkspaceService),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

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
	case resource.Event[*module.Module]:
		switch msg.Type {
		case resource.CreatedEvent:
			//		cmds = append(cmds, tui.NavigateTo(tui.ModuleKind, &msg.Payload.Resource))
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
			return tui.BodyResizeMsg{Width: m.viewWidth(), Height: m.viewHeight()}
		}
	case tea.KeyMsg:
		// Pressing any key makes any info/error message in the footer disappear
		m.info = ""
		m.err = nil

		switch {
		case key.Matches(msg, keys.Global.Quit):
			// ctrl-c quits the app
			return m, tea.Quit
		case key.Matches(msg, keys.Global.Escape):
			// <esc> closes help or goes back to last page
			if m.showHelp {
				m.showHelp = false
			} else {
				m.goBack()
			}
		case key.Matches(msg, keys.Global.Help):
			// '?' toggles help
			m.showHelp = !m.showHelp
		case key.Matches(msg, keys.Global.Logs):
			// 'l' shows logs
			return m, tui.NavigateTo(tui.LogsKind)
		case key.Matches(msg, keys.Global.Modules):
			// 'm' lists all modules
			return m, tui.NavigateTo(tui.ModuleListKind)
		case key.Matches(msg, keys.Global.Workspaces):
			// 'W' lists all workspaces
			return m, tui.NavigateTo(tui.WorkspaceListKind)
		case key.Matches(msg, keys.Global.Runs):
			// 'R' lists all runs
			return m, tui.NavigateTo(tui.RunListKind)
		case key.Matches(msg, keys.Global.Tasks):
			// 'T' lists all tasks
			return m, tui.NavigateTo(tui.TaskListKind)
		default:
			// Send other keys to current model.
			cmd := m.updateCurrent(msg)
			return m, cmd
		}
	case tui.NavigationMsg:
		created, err := m.setCurrent(msg.Page)
		if err != nil {
			return m, tui.ReportError(err, "setting current page")
		}
		if created {
			cmds = append(cmds, m.currentModel().Init())
		}
	case tui.CreatedRunsMsg:
		cmd, m.info, m.err = handleCreatedRunsMsg(msg)
		cmds = append(cmds, cmd)
	case tui.CreatedTasksMsg:
		cmd, m.info, m.err = handleCreatedTasksMsg(msg)
		cmds = append(cmds, cmd)
	case tui.CompletedTasksMsg:
		m.info, m.err = handleCompletedTasksMsg(msg)
	case tui.ErrorMsg:
		if msg.Error != nil {
			err := msg.Error
			msg := fmt.Sprintf(msg.Message, msg.Args...)

			// Both print error in footer as well as log it.
			m.err = fmt.Errorf("%s: %w", msg, err)
			slog.Error(msg, "error", err)
		}
	case tui.InfoMsg:
		m.info = string(msg)
	default:
		// Send remaining msg types to all cached models
		cmds = append(cmds, m.cache.updateAll(msg)...)
	}
	return m, tea.Batch(cmds...)
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

	workdirIcon = tui.Bold.Copy().
			Foreground(tui.Pink).
			Margin(0, 2, 0, 1).
			Render("🗀")
	versionIcon = tui.Regular.Copy().
			Foreground(tui.Pink).
			Margin(0, 2, 0, 1).
			Render("ⓥ")
	workdirStyle = tui.Regular.Foreground(tui.LightGrey)
	versionStyle = tui.Regular.Foreground(tui.LightGrey)
)

func (m model) View() string {
	var (
		content           string
		shortHelpBindings []key.Binding
	)

	if m.showHelp {
		content = lipgloss.NewStyle().
			Margin(1).
			Render(
				fullHelpView(
					m.currentModel().HelpBindings(),
					keys.KeyMapToSlice(keys.Global),
					keys.KeyMapToSlice(keys.Navigation),
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
		// Within the header, show the model bindings first, and then the
		// general key bindings. The navigation bindings are only visible in the
		// full help.
		shortHelpBindings = append(
			m.currentModel().HelpBindings(),
			keys.KeyMapToSlice(keys.Global)...,
		)
	}

	// Render global static info in top left corner
	globalStatic := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Left, workdirIcon, workdirStyle.Render(m.workdir)),
		lipgloss.JoinHorizontal(lipgloss.Left, versionIcon, versionStyle.Render(version.Version)),
	)

	// Render help bindings in between version and logo. Set its available width
	// to the width of the terminal minus the width of the global static info,
	// the width of the logo, and the width of its margins.
	shortHelpWidth := m.width - tui.Width(globalStatic) - logoWidth - 6
	shortHelp := lipgloss.NewStyle().
		Margin(0, 2, 0, 4).
		Width(shortHelpWidth).
		Render(shortHelpView(shortHelpBindings, shortHelpWidth))

	// Render title.
	title := lipgloss.NewStyle().
		Margin(0, 1).
		Render(m.currentModel().Title())

	// Optionally render page id and/or status to the right side of title
	pageIDAndStatusStyle := tui.Regular.
		Margin(0, 1).
		Width(m.width - tui.Width(title) - 2).
		Align(lipgloss.Right)
	var (
		pageID     string
		pageStatus string
	)
	if identifiable, ok := m.currentModel().(tui.ModelID); ok {
		pageID = tui.Regular.Copy().Padding(0, 0, 0, 0).Render(identifiable.ID())
	}
	if statusable, ok := m.currentModel().(tui.ModelStatus); ok {
		pageStatus = tui.Padded.Copy().Render(statusable.Status())
	}
	pageIDAndStatus := pageIDAndStatusStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Left, pageStatus, pageID),
	)
	title = lipgloss.JoinHorizontal(lipgloss.Left, title, pageIDAndStatus)

	// Global-level info goes in the bottom right corner in the footer.
	metadata := tui.Padded.Copy().Render(
		fmt.Sprintf("%d/%d tasks", m.tasks.Counter(), 32),
	)

	// Render any info/error message to be shown in the bottom left corner in
	// the footer, using whatever space is remaining to the left of the
	// metadata.
	var footerMsg string
	if m.err != nil {
		footerMsg = tui.Padded.Copy().
			Foreground(tui.Red).
			Render("Error: " + m.err.Error())
	} else if m.info != "" {
		footerMsg = tui.Padded.Copy().
			Foreground(tui.Black).
			Render(m.info)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		// header
		lipgloss.NewStyle().
			Height(headerHeight).
			Render(
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					// global static info
					globalStatic,
					// help
					shortHelp,
					// logo
					renderedLogo,
				),
			),
		// title
		lipgloss.NewStyle().
			// Prohibit overflowing title wrapping to another line.
			MaxHeight(1).
			Inline(true).
			Width(m.width).
			// Prefix title with a space to add margin (Inline() doesn't permit
			// using Margin()).
			Render(title),
		// horizontal rule
		strings.Repeat("─", m.width),
		// content
		lipgloss.NewStyle().
			Height(m.viewHeight()).
			Render(content),
		// horizontal rule
		strings.Repeat("─", m.width),
		// footer
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			// info/error message
			tui.Regular.
				Inline(true).
				MaxWidth(m.width-tui.Width(metadata)).
				Width(m.width-tui.Width(metadata)).
				Render(footerMsg),
			// pagination
			metadata,
		),
	)
}

// viewHeight retrieves the height available beneath the header and breadcrumbs,
// and the message footer.
func (m model) viewHeight() int {
	// Take total terminal height and subtract the height of the header, the
	// title, the horizontal rule under the title, and then in the footer, the
	// horizontal rule and the message underneath.
	return m.height - headerHeight - breadcrumbsHeight - 2*horizontalRuleHeight - messageFooterHeight
}

// viewWidth retrieves the width available within the main view
func (m model) viewWidth() int {
	return m.width
}
