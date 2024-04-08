package top

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/version"
)

type model struct {
	WorkspaceService tui.WorkspaceService

	*navigator

	width  int
	height int

	showHelp bool

	showQuitPrompt bool
	quitPrompt     textinput.Model

	// Either an error or an informational message is rendered in the footer.
	err  error
	info string

	tasks    tui.TaskService
	spinner  *spinner.Model
	maxTasks int

	dump *os.File

	workdir string
}

type Options struct {
	ModuleService    tui.ModuleService
	WorkspaceService tui.WorkspaceService
	StateService     tui.StateService
	RunService       tui.RunService
	TaskService      tui.TaskService

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
		maxTasks:         opts.MaxTasks,
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

	if m.showQuitPrompt {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.Global.Quit):
				// pressing ctrl-c again quits the app
				return m, tea.Quit
			case key.Matches(msg, localKeys.Yes):
				// 'y' quits the app
				return m, tea.Quit
			default:
				// any other key closes the prompt and returns to the app
				m.showQuitPrompt = false
				m.info = "canceled quitting pug"
			}
		}
		return m, cmd
	}

	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height

		// Inform navigator of new dimenisions for when it builds new models
		m.navigator.width = m.viewWidth()
		m.navigator.height = m.viewHeight()

		// amend msg to account for header etc, and forward below to all cached
		// models.
		msg = tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Pressing any key makes any info/error message in the footer disappear
		m.info = ""
		m.err = nil

		switch {
		case key.Matches(msg, keys.Global.Quit):
			// ctrl-c quits the app, but not before prompting the user for
			// comfirmation.
			m.quitPrompt = textinput.New()
			m.quitPrompt.Prompt = ""
			m.quitPrompt.Focus()
			m.showQuitPrompt = true
			return m, textinput.Blink
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
	versionIcon = tui.Bold.Copy().
			Foreground(tui.Pink).
			Margin(0, 2, 0, 1).
			Render("ⓥ")
	workdirStyle = tui.Regular.Copy()
	versionStyle = tui.Regular.Copy()
)

func (m model) View() string {
	var (
		content           string
		shortHelpBindings []key.Binding
	)

	var currentHelpBindings []key.Binding
	if bindings, ok := m.currentModel().(tui.ModelHelpBindings); ok {
		currentHelpBindings = bindings.HelpBindings()
	}

	if m.showHelp {
		content = lipgloss.NewStyle().
			Margin(1).
			Render(
				fullHelpView(
					currentHelpBindings,
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
	} else if m.showQuitPrompt {
		content = lipgloss.NewStyle().
			Margin(0, 1).
			Render(fmt.Sprintf("Quit pug? (y/N): %s", m.quitPrompt.View()))
	} else {
		content = m.currentModel().View()
		shortHelpBindings = append(
			currentHelpBindings,
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

	// Render page title line
	var (
		pageTitle  string
		pageID     string
		pageStatus string
	)
	if titled, ok := m.currentModel().(tui.ModelTitle); ok {
		pageTitle = tui.Regular.Copy().Margin(0, 1).Render(titled.Title())
	}

	// Optionally render page id and/or status to the right side of title
	pageIDAndStatusStyle := tui.Regular.
		Margin(0, 1).
		Width(m.width - tui.Width(pageTitle) - 2).
		Align(lipgloss.Right)
	if identifiable, ok := m.currentModel().(tui.ModelID); ok {
		pageID = tui.Regular.Copy().Padding(0, 0, 0, 0).Render(identifiable.ID())
	}
	if statusable, ok := m.currentModel().(tui.ModelStatus); ok {
		pageStatus = tui.Padded.Copy().Render(statusable.Status())
	}
	pageIDAndStatus := pageIDAndStatusStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Left, pageStatus, pageID),
	)

	// Stitch together page title line, and id and status to the right
	pageTitleLine := lipgloss.JoinHorizontal(lipgloss.Left, pageTitle, pageIDAndStatus)

	// Global-level info goes in the bottom right corner in the footer.
	metadata := tui.Padded.Copy().Render(
		fmt.Sprintf("%d/%d tasks", m.tasks.Counter(), m.maxTasks),
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
			Render(pageTitleLine),
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
