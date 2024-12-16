package top

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/leg100/pug/internal/app"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	tuimodule "github.com/leg100/pug/internal/tui/module"
	"github.com/leg100/pug/internal/version"
)

// pug is in one of several modes, which alter how all messages are handled.
type mode int

const (
	normalMode mode = iota // default
	promptMode             // confirm prompt is visible and taking input
	filterMode             // filter is visible and taking input
)

type model struct {
	*tui.PaneManager

	makers   map[tui.Kind]tui.Maker
	modules  *module.Service
	width    int
	height   int
	mode     mode
	showHelp bool
	prompt   *tui.Prompt
	dump     *os.File
	workdir  string
	err      error
	info     string
	tasks    *task.Service
	spinner  *spinner.Model
	spinning bool
	maxTasks int
}

func newModel(cfg app.Config, app *app.App) (model, error) {
	var dump *os.File
	if cfg.Debug {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return model{}, err
		}
	}

	// Work-around for
	// https://github.com/charmbracelet/bubbletea/issues/1036#issuecomment-2158563056
	_ = lipgloss.HasDarkBackground()

	helpers := &tui.Helpers{
		Modules:    app.Modules,
		Workspaces: app.Workspaces,
		Plans:      app.Plans,
		States:     app.States,
		Tasks:      app.Tasks,
		Logger:     app.Logger,
	}
	spinner := spinner.New(spinner.WithSpinner(spinner.Line))
	makers := makeMakers(cfg, app, &spinner, helpers)

	m := model{
		PaneManager: tui.NewPaneManager(makers),
		makers:      makers,
		modules:     app.Modules,
		spinner:     &spinner,
		tasks:       app.Tasks,
		maxTasks:    cfg.MaxTasks,
		dump:        dump,
		workdir:     cfg.Workdir.PrettyString(),
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.PaneManager.Init(),
		tuimodule.ReloadModules(true, m.modules),
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
		//_ = m.updateCurrent(msg)
		if m.spinning {
			// Continue spinning spinner.
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tui.PromptMsg:
		// Enable prompt widget
		m.mode = promptMode
		var blink tea.Cmd
		m.prompt, blink = tui.NewPrompt(msg)
		// Send out message to panes to resize themselves to make room for the prompt above it.
		_ = m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
		return m, tea.Batch(cmd, blink)
	case tea.KeyMsg:
		// Pressing any key makes any info/error message in the footer disappear
		m.info = ""
		m.err = nil

		switch m.mode {
		case promptMode:
			closePrompt, cmd := m.prompt.HandleKey(msg)
			if closePrompt {
				// Send message to panes to resize themselves to expand back
				// into space occupied by prompt.
				m.mode = normalMode
				_ = m.PaneManager.Update(tea.WindowSizeMsg{
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
				cmd = m.ActiveModel().Update(tui.FilterBlurMsg{})
				return m, cmd
			case key.Matches(msg, keys.Filter.Blur):
				// Switch back to normal mode, and send message to current model
				// to blur the filter widget
				m.mode = normalMode
				cmd = m.ActiveModel().Update(tui.FilterBlurMsg{})
				return m, cmd
			case key.Matches(msg, keys.Filter.Close):
				// Switch back to normal mode, and send message to current model
				// to close the filter widget
				m.mode = normalMode
				cmd = m.ActiveModel().Update(tui.FilterCloseMsg{})
				return m, cmd
			default:
				// Wrap key message in a filter key message and send to current
				// model.
				cmd = m.ActiveModel().Update(tui.FilterKeyMsg(msg))
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, keys.Global.Quit):
			// ctrl-c quits the app, but not before prompting the user for
			// confirmation.
			return m, tui.YesNoPrompt("Quit pug?", tea.Quit)
		case key.Matches(msg, keys.Global.Suspend):
			// ctrl-z suspends the app
			return m, tea.Suspend
		case key.Matches(msg, keys.Global.Help):
			// '?' toggles help widget
			m.showHelp = !m.showHelp
			// Help widget takes up space so update panes' dimensions
			m.PaneManager.Update(tea.WindowSizeMsg{
				Height: m.viewHeight(),
				Width:  m.viewWidth(),
			})
		case key.Matches(msg, keys.Global.Filter):
			// '/' enables filter mode if the current model indicates it
			// supports it, which it does so by sending back a non-nil command.
			if cmd = m.ActiveModel().Update(tui.FilterFocusReqMsg{}); cmd != nil {
				m.mode = filterMode
			}
			return m, cmd
		case key.Matches(msg, keys.Global.Explorer):
			return m, tui.NavigateTo(tui.ExplorerKind, tui.WithPosition(tui.LeftPane))
		case key.Matches(msg, keys.Global.TaskGroups):
			// list all taskgroups
			return m, tui.NavigateTo(tui.TaskGroupListKind)
		case key.Matches(msg, keys.Global.Logs):
			return m, tui.NavigateTo(tui.LogListKind)
		case key.Matches(msg, keys.Global.Tasks):
			return m, tui.NavigateTo(tui.TaskListKind)
		default:
			// Send all other keys to panes.
			if cmd := m.PaneManager.Update(msg); cmd != nil {
				return m, cmd
			}
			// If pane manager doesn't respond with a command, then send key to
			// any updateable model makers; first one to respond with a command
			// wins.
			for _, maker := range m.makers {
				if updateable, ok := maker.(updateableMaker); ok {
					if cmd := updateable.Update(msg); cmd != nil {
						return m, cmd
					}
				}
			}
			return m, nil
		}
	case tui.ErrorMsg:
		m.err = error(msg)
	case tui.InfoMsg:
		m.info = string(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
	case cursor.BlinkMsg:
		// Send blink message to prompt if in prompt mode otherwise forward it
		// to the active pane to handle.
		if m.mode == promptMode {
			cmd = m.prompt.HandleBlink(msg)
		} else {
			cmd = m.ActiveModel().Update(msg)
		}
		return m, cmd
	default:
		// Send remaining msg types to pane manager to route accordingly.
		cmds = append(cmds, m.PaneManager.Update(msg))
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Start composing vertical stack of components that fill entire terminal.
	var components []string

	// Add prompt if in prompt mode.
	if m.mode == promptMode {
		components = append(components, m.prompt.View(m.width))
	}
	// Add panes
	components = append(components, lipgloss.NewStyle().
		Height(m.viewHeight()).
		Width(m.viewWidth()).
		Render(m.PaneManager.View()),
	)
	// Add help if enabled
	if m.showHelp {
		components = append(components, m.help())
	}
	// Compose footer
	footer := tui.Padded.Background(tui.Grey).Foreground(tui.White).Render("? help")
	if m.err != nil {
		footer += tui.Padded.
			Bold(true).
			Background(tui.Red).
			Foreground(tui.White).
			Render("Error:")
		footer += tui.Regular.Padding(0, 1, 0, 0).
			Background(tui.Red).
			Foreground(tui.White).
			Render(m.err.Error())
	} else if m.info != "" {
		footer += tui.Padded.
			Foreground(tui.Black).
			Background(tui.EvenLighterGrey).
			Render(m.info)
	}
	pug := tui.Padded.Background(tui.LightGrey).Foreground(tui.White).Render("PUG")
	version := tui.Padded.Background(tui.DarkGrey).Foreground(tui.White).Render(version.Version)
	// Fill in left over space with background color
	leftover := m.width - tui.Width(footer) - tui.Width(pug) - tui.Width(version)
	footer += tui.Regular.Width(leftover).Background(tui.EvenLighterGrey).Render()
	footer += version
	footer += pug
	// Add footer
	components = append(components, tui.Regular.
		Inline(true).
		MaxWidth(m.width).
		Width(m.width).
		Render(footer),
	)
	return lipgloss.JoinVertical(lipgloss.Top, components...)
}

// viewHeight returns the height available to the panes
//
// TODO: rename contentHeight
func (m model) viewHeight() int {
	vh := m.height - tui.FooterHeight
	if m.mode == promptMode {
		vh -= tui.PromptHeight
	}
	if m.showHelp {
		vh -= tui.HelpWidgetHeight
	}
	return max(tui.MinContentHeight, vh)
}

// viewWidth retrieves the width available within the main view
//
// TODO: rename contentWidth
func (m model) viewWidth() int {
	return max(tui.MinContentWidth, m.width)
}

var (
	helpKeyStyle  = tui.Bold.Foreground(tui.HelpKey).Margin(0, 1, 0, 0)
	helpDescStyle = tui.Regular.Foreground(tui.HelpDesc)
)

// help renders key bindings
func (m model) help() string {
	// Compile list of bindings to render
	bindings := []key.Binding{keys.Global.Help, keys.Global.Quit}
	switch m.mode {
	case promptMode:
		bindings = append(bindings, m.prompt.HelpBindings()...)
	case filterMode:
		bindings = append(bindings, keys.KeyMapToSlice(keys.Filter)...)
	default:
		if model, ok := m.ActiveModel().(tui.ModelHelpBindings); ok {
			bindings = append(bindings, model.HelpBindings()...)
		}
	}
	bindings = append(bindings, keys.KeyMapToSlice(keys.Global)...)
	bindings = append(bindings, keys.KeyMapToSlice(keys.Navigation)...)
	bindings = removeDuplicateBindings(bindings)

	// Enumerate through each group of bindings, populating a series of
	// pairs of columns, one for keys, one for descriptions
	var (
		pairs []string
		width int
		// Subtract 2 to accommodate borders
		rows = tui.HelpWidgetHeight - 2
	)
	for i := 0; i < len(bindings); i += rows {
		var (
			keys  []string
			descs []string
		)
		for j := i; j < min(i+rows, len(bindings)); j++ {
			keys = append(keys, helpKeyStyle.Render(bindings[j].Help().Key))
			descs = append(descs, helpDescStyle.Render(bindings[j].Help().Desc))
		}
		// Render pair of columns; beyond the first pair, render a three space
		// left margin, in order to visually separate the pairs.
		var cols []string
		if len(pairs) > 0 {
			cols = []string{"   "}
		}
		cols = append(cols,
			strings.Join(keys, "\n"),
			strings.Join(descs, "\n"),
		)

		pair := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
		// check whether it exceeds the maximum width avail (the width of the
		// terminal, subtracting 2 for the borders).
		width += lipgloss.Width(pair)
		if width > m.width-2 {
			break
		}
		pairs = append(pairs, pair)
	}
	// Join pairs of columns and enclose in a border
	content := lipgloss.JoinHorizontal(lipgloss.Top, pairs...)
	return tui.Border.Height(rows).Width(m.width - 2).Render(content)
}

// removeDuplicateBindings removes duplicate bindings from a list of bindings. A
// binding is deemed a duplicate if another binding has the same list of keys.
func removeDuplicateBindings(bindings []key.Binding) []key.Binding {
	seen := make(map[string]struct{})
	var i int
	for _, b := range bindings {
		key := strings.Join(b.Keys(), " ")
		if _, ok := seen[key]; ok {
			// duplicate, skip
			continue
		}
		seen[key] = struct{}{}
		bindings[i] = b
		i++
	}
	return bindings[:i]
}
