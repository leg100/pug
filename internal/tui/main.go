package tui

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

const defaultState = moduleListState

type main struct {
	current Page

	models map[Page]Model

	width  int
	height int

	// status contains extraordinary info, e.g. errors, warnings
	status   string
	messages *slog.Logger
}

type Options struct {
	TaskService      *task.Service
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	RunService       *run.Service
	Workdir          string
}

func New(opts Options) (main, error) {
	mm, err := newModuleListModel(opts.ModuleService, opts.Workdir)
	if err != nil {
		return main{}, err
	}

	f, err := os.OpenFile("messages.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return main{}, err
	}
	messages := slog.New(slog.NewTextHandler(f, nil))

	return main{
		models: map[Page]Model{
			moduleListState: mm,
			logsState:       newLogs(),
			helpState:       newHelp(defaultState),
		},
		current:  defaultState,
		messages: messages,
	}, nil
}

func (m main) Init() tea.Cmd {
	inits := make([]tea.Cmd, len(m.models))
	for i, mod := range maps.Values(m.models) {
		inits[i] = mod.Init()
	}
	return tea.Batch(inits...)
}

func (m main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if _, ok := msg.(logMsg); !ok {
		m.messages.Info(fmt.Sprintf("%v", msg), "type", reflect.TypeOf(msg))
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.resizeCmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit
		default:
			// send key to current model
			updated, cmd := m.models[m.current].Update(msg)
			m.models[m.current] = updated
			// then relay key as global key, along with any command the current
			// model returns
			return m, tea.Batch(
				cmd,
				cmdHandler(globalKeyMsg{
					Current: m.current,
					KeyMsg:  msg,
				}),
			)
		}
	case navigationMsg:
		m.current = msg.To
		if msg.Model != nil {
			m.models[m.current] = msg.Model
			cmds = append(cmds, msg.Model.Init(), m.resizeCmd)
		}
	}
	// relay messages to all models and update accordingly
	for k, mod := range m.models {
		updated, cmd := mod.Update(msg)
		m.models[k] = updated
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m main) resizeCmd() tea.Msg {
	return viewSizeMsg{Width: m.viewWidth(), Height: m.viewHeight()}
}

// viewHeight retrieves the height available within the main view
func (m main) viewHeight() int {
	// hardcode height adjustment for performance reasons:
	// heading: 3
	// hr: 1
	// title: 1
	// hr: 1
	return m.height - 4
}

// viewWidth retrieves the width available within the main view
func (m main) viewWidth() int {
	return m.width
}

func (m main) View() string {
	current := m.models[m.current]
	title := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(DarkGrey)).
		Foreground(White).
		Padding(0, 1).
		Render(current.Title())
	titleWidth := lipgloss.Width(title)

	rows := []string{
		m.header(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			"─",
			title,
			strings.Repeat("─", max(0, m.width-titleWidth)),
		),
		Regular.Copy().
			Height(m.viewHeight()).
			Width(m.viewWidth()).
			Render(current.View()),
	}
	return lipgloss.JoinVertical(lipgloss.Top, rows...)
}

var (
	logo = strings.Join([]string{
		"▄▄▄ ▄ ▄ ▄▄▄",
		"█▄█ █ █ █ ▄",
		"▀   ▀▀▀ ▀▀▀",
	}, "\n")
	renderedLogo = Bold.
			Copy().
			Padding(0, 1).
			Foreground(Pink).
			Render(logo)
	logoWidth = lipgloss.Width(renderedLogo)
)

func (m main) header() string {
	help := lipgloss.NewStyle().
		Width(m.width - logoWidth).
		Render(RenderShort(m.current))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		help,
		renderedLogo,
	)
}
