package task

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/muesli/reflow/wordwrap"
)

type Maker struct {
	TaskService tui.TaskService
	Spinner     *spinner.Model

	// If IsRunTab is true then Maker makes task models that are a tab within
	// the run model.
	IsRunTab bool

	Helpers *tui.Helpers
}

func (mm *Maker) Make(tr resource.Resource, width, height int) (tui.Model, error) {
	task, err := mm.TaskService.Get(tr.ID)
	if err != nil {
		return model{}, err
	}

	m := model{
		svc:      mm.TaskService,
		task:     task,
		output:   task.NewReader(),
		spinner:  mm.Spinner,
		isRunTab: mm.IsRunTab,
		// read upto 1kb at a time
		buf:     make([]byte, 1024),
		width:   width,
		height:  height,
		helpers: mm.Helpers,
	}

	m.viewport = viewport.New(0, 0)
	m.viewport.HighPerformanceRendering = false
	m.setViewportDimensions(width, height)

	return m, nil
}

type model struct {
	svc  tui.TaskService
	task *task.Task

	output   io.Reader
	buf      []byte
	content  string
	isRunTab bool

	viewport viewport.Model
	spinner  *spinner.Model

	width  int
	height int

	showInfo bool

	helpers *tui.Helpers
}

func (m model) Init() tea.Cmd {
	return m.getOutput
}

func (m model) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Info):
			// 'i' toggles showing task info
			m.showInfo = !m.showInfo
		case key.Matches(msg, keys.Common.Cancel):
			return m, tui.CreateTasks("cancel", m.svc.Cancel, m.task.ID)
			// TODO: retry
		case key.Matches(msg, keys.Common.Apply):

			return m, tui.CreateTasks("cancel", m.svc.Cancel, m.task.ID)
			// TODO: retry
		}
	case outputMsg:
		if msg.taskID != m.task.ID {
			return m, nil
		}
		// isChild is true when this msg is for a task model that is a child of a
		// run model, i.e. a tab. Without this flag, output would be duplicated in
		// both the tab and on the generic task view.
		if msg.isRunTab != m.isRunTab {
			return m, nil
		}
		m.content += msg.output
		m.content = wordwrap.String(m.content, m.width)
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		if !msg.eof {
			cmds = append(cmds, m.getOutput)
		}
	case resource.Event[*task.Task]:
		if msg.Payload.ID != m.task.ID {
			// Ignore event for different task.
			return m, nil
		}
		m.task = msg.Payload
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.setViewportDimensions(msg.Width, msg.Height)
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	heading := tui.Bold.Render("Task")
	cmd := tui.Regular.Copy().Foreground(tui.Green).Render(m.task.CommandString())
	crumbs := m.helpers.Breadcrumbs("", *m.task.Parent)
	return fmt.Sprintf("%s{%s}%s", heading, cmd, crumbs)
}

func (m model) ID() string {
	return m.task.String()
}

const (
	// paginationWidth is the width of the pagination section to the right of
	// the viewport
	paginationWidth = 10
	// viewportMarginsWidth is the total width of the margins to the left and
	// right of the viewport
	viewportMarginsWidth = 2
)

func (m *model) setViewportDimensions(width, height int) {
	// minusWidth is the width to subtract from that available to the viewport
	minusWidth := paginationWidth - viewportMarginsWidth

	// width is the available to the viewport
	width = max(0, width-minusWidth)

	m.viewport.Width = width
	m.viewport.Height = height
}

// View renders the viewport
func (m model) View() string {
	if m.showInfo {
		return strings.Join(m.task.Env, " ")
	}

	viewport := tui.Regular.Copy().
		Margin(0, 1).
		MaxWidth(m.viewport.Width).
		Render(m.viewport.View())

	// pagination info container occupies a fixed width section to the right of
	// the viewport.
	scrollPercent := tui.Regular.Copy().
		Background(lipgloss.Color("253")).
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	pagination := tui.Regular.Copy().
		Margin(0, 1).
		Height(m.height).
		Width(paginationWidth - 2).
		AlignVertical(lipgloss.Bottom).
		Render(scrollPercent)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		viewport,
		pagination,
	)
}

func (m model) TabStatus() string {
	switch m.task.State {
	case task.Running:
		return m.spinner.View()
	case task.Exited:
		return "✓"
	case task.Errored:
		return "✗"
	}
	return "+"
}

func (m model) Status() string {
	var color lipgloss.Color

	switch m.task.State {
	case task.Pending:
		color = tui.Grey
	case task.Queued:
		color = tui.Orange
	case task.Running:
		color = tui.Blue
	case task.Exited:
		color = tui.GreenBlue
	case task.Errored:
		color = tui.Red
	}
	return tui.Regular.Copy().Foreground(color).Render(string(m.task.State))
}

func (m model) HelpBindings() (bindings []key.Binding) {
	// TODO: filter keys depending upon current task.
	return []key.Binding{
		keys.Common.Plan,
		keys.Common.Apply,
		keys.Common.Cancel,
	}
}

func (m model) getOutput() tea.Msg {
	msg := outputMsg{taskID: m.task.ID, isRunTab: m.isRunTab}

	n, err := m.output.Read(m.buf)
	if err == io.EOF {
		msg.eof = true
	} else if err != nil {
		return tui.NewErrorMsg(err, "reading task output")
	}
	msg.output = string(m.buf[:n])
	return msg
}

type outputMsg struct {
	isRunTab bool
	taskID   resource.ID
	output   string
	eof      bool
}
