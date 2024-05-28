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
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/module"
	tuitask "github.com/leg100/pug/internal/tui/task"
	"github.com/leg100/pug/internal/version"
)

// pug is in one of several modes, which alter how all messages are handled.
type mode int

const (
	normalMode mode = iota // default
	helpMode               // help is visible
	promptMode             // confirm prompt is visible and taking input
	filterMode             // filter is visible and taking input
)

type model struct {
	ModuleService tui.ModuleService

	*navigator

	width  int
	height int

	mode mode

	prompt *tui.Prompt

	// Either an error or an informational message is rendered in the footer.
	err  error
	info string

	tasks    tui.TaskService
	spinner  *spinner.Model
	spinning bool
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
	Workdir   internal.Workdir
	FirstPage string
	MaxTasks  int
	Debug     bool
}

// New constructs the top-level TUI model.
func New(opts Options) (model, error) {
	var dump *os.File
	if opts.Debug {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return model{}, err
		}
	}

	spinner := spinner.New(spinner.WithSpinner(spinner.Line))

	navigator, err := newNavigator(opts, &spinner)
	if err != nil {
		return model{}, err
	}

	m := model{
		ModuleService: opts.ModuleService,
		navigator:     navigator,
		spinner:       &spinner,
		tasks:         opts.TaskService,
		maxTasks:      opts.MaxTasks,
		dump:          dump,
		workdir:       opts.Workdir.PrettyString(),
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.currentModel().Init(),
		module.ReloadModules(true, m.ModuleService),
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
			if !m.spinning {
				// There are tasks running and the spinner isn't spinning, so start
				// the spinner.
				m.spinning = true
				cmds = append(cmds, m.spinner.Tick)
			}
		} else {
			// No tasks are running so stop spinner
			m.spinning = false
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		*m.spinner, cmd = m.spinner.Update(msg)
		_ = m.updateCurrent(msg)
		if m.spinning {
			// Continue spinning spinner.
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tui.FilterFocusAckMsg:
		// The filter widget has acknowledged the focus request, so we can now
		// enable filter mode.
		m.mode = filterMode
	case tui.PromptMsg:
		// Enable prompt widget
		m.mode = promptMode
		var blink tea.Cmd
		m.prompt, blink = tui.NewPrompt(msg)
		// Send out message to current model to resize itself to make room for
		// the prompt above it.
		cmd := m.updateCurrent(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
		return m, tea.Batch(cmd, blink)
	case tea.KeyMsg:
		// Pressing any key makes any info/error message in the footer disappear
		m.info = ""
		m.err = nil

		switch m.mode {
		case helpMode:
			switch {
			case key.Matches(msg, keys.Global.Quit):
				// Let quit key handler below handle this
				break
			case key.Matches(msg, keys.Global.Help, keys.Global.Back):
				// Exit help
				m.mode = normalMode
				return m, nil
			default:
				// Any other key is ignored
				return m, nil
			}
		case promptMode:
			closePrompt, cmd := m.prompt.HandleKey(msg)
			if closePrompt {
				// Send message to current model to resize itself to expand back
				// into space occupied by prompt.
				m.mode = normalMode
				_ = m.updateCurrent(tea.WindowSizeMsg{
					Height: m.viewHeight(),
					Width:  m.viewWidth(),
				})
			}
			return m, cmd
		case filterMode:
			switch {
			case key.Matches(msg, keys.Global.Quit):
				// Allow user to quit app whilst in filter mode. In this case,
				// switch back to normal mode, blur the filter widget, and let
				// the key handler below handle the quit action.
				m.mode = normalMode
				_ = m.updateCurrent(tui.FilterBlurMsg{})
			case key.Matches(msg, keys.Filter.Blur):
				// Switch back to normal mode, and send message to current model
				// to blur the filter widget
				m.mode = normalMode
				_ = m.updateCurrent(tui.FilterBlurMsg{})
				return m, nil
			case key.Matches(msg, keys.Filter.Close):
				// Switch back to normal mode, and send message to current model
				// to close the filter widget
				m.mode = normalMode
				_ = m.updateCurrent(tui.FilterCloseMsg{})
				return m, nil
			default:
				// Wrap key message in a filter key message and send to current
				// model.
				cmd = m.updateCurrent(tui.FilterKeyMsg(msg))
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, keys.Global.Quit):
			// ctrl-c quits the app, but not before prompting the user for
			// confirmation.
			return m, tui.YesNoPrompt("Quit pug?", tea.Quit)
		case key.Matches(msg, keys.Global.Back):
			// <esc> goes back to last page
			m.goBack()
		case key.Matches(msg, keys.Global.Help):
			// '?' enables help mode
			m.mode = helpMode
		case key.Matches(msg, keys.Global.Filter):
			// '/' enables filter mode, but only if the current model
			// acknowledges the message.
			cmd = m.updateCurrent(tui.FilterFocusReqMsg{})
			return m, cmd
		case key.Matches(msg, keys.Global.Logs):
			// 'l' shows logs
			return m, tui.NavigateTo(tui.LogListKind)
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
		if msg.Tab != "" {
			cmds = append(cmds, m.updateCurrent(tui.SetActiveTabMsg(msg.Tab)))
		}
	case tuitask.CreatedTasksMsg:
		cmd, m.info, m.err = tuitask.HandleCreatedTasks(msg)
		cmds = append(cmds, cmd)
	case tuitask.CompletedTasksMsg:
		m.info, m.err = tuitask.HandleCompletedTasks(msg)
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Inform navigator of new dimensions for when it builds new models
		m.navigator.width = m.viewWidth()
		m.navigator.height = m.viewHeight()

		// amend msg to account for header etc, and forward to all cached
		// models.
		_ = m.cache.updateAll(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
	default:
		// Send remaining msg types to all cached models
		cmds = append(cmds, m.cache.updateAll(msg)...)

		// Send message to the prompt too if in prompt mode (most likely a
		// blink message)
		if m.mode == promptMode {
			cmds = append(cmds, m.prompt.HandleBlink(msg))
		}
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
		currentHelpBindings = tui.RemoveDuplicateBindings(currentHelpBindings)
	}

	switch m.mode {
	case helpMode:
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
	case promptMode:
		content = m.currentModel().View()
		shortHelpBindings = m.prompt.HelpBindings()
	case filterMode:
		content = m.currentModel().View()
		shortHelpBindings = keys.KeyMapToSlice(keys.Filter)
	default:
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

	// Vertical stack of components that make up the rendered view.
	components := []string{
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
	}

	if m.mode == promptMode {
		components = append(components,
			tui.Regular.Margin(0, 1).Render(m.prompt.View()),
			// horizontal rule
			strings.Repeat("─", m.width),
		)
	}

	components = append(components,
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

	return lipgloss.JoinVertical(lipgloss.Top, components...)
}

const promptHeight = 2

// viewHeight returns the height available to the current model (subordinate to
// the top model).
func (m model) viewHeight() int {
	vh := m.height - headerHeight - breadcrumbsHeight - 2*horizontalRuleHeight - messageFooterHeight
	if m.mode == promptMode {
		vh -= promptHeight
	}
	return vh
}

// viewWidth retrieves the width available within the main view
func (m model) viewWidth() int {
	return m.width
}
