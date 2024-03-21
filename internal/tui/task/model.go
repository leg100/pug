package task

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/muesli/reflow/wordwrap"
)

type Maker struct {
	TaskService *task.Service

	// If IsRunTab is true then Maker makes task models that are a tab within
	// the run model.
	IsRunTab bool
}

func (mm *Maker) Make(tr resource.Resource, width, height int) (tui.Model, error) {
	task, err := mm.TaskService.Get(tr.ID())
	if err != nil {
		return model{}, err
	}

	m := model{
		svc:      mm.TaskService,
		task:     task,
		output:   task.NewReader(),
		viewport: viewport.New(0, 0),
		isRunTab: mm.IsRunTab,
		// read upto 1kb at a time
		buf:    make([]byte, 1024),
		width:  width,
		height: height,
	}

	m.setViewportDimensions(width, height)

	return m, nil
}

type model struct {
	svc  *task.Service
	task *task.Task

	output   io.Reader
	buf      []byte
	content  string
	viewport viewport.Model
	isRunTab bool

	width  int
	height int
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
		case key.Matches(msg, keys.Common.Cancel):
			return m, TaskCmd(m.svc.Cancel, m.task.ID())
			// TODO: retry
		case key.Matches(msg, keys.Common.Apply):

			return m, TaskCmd(m.svc.Cancel, m.task.ID())
			// TODO: retry
		}
	case outputMsg:
		if msg.taskID != m.task.ID() {
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
		if msg.Payload.ID() != m.task.ID() {
			// Ignore event for different task.
			return m, nil
		}
		m.task = msg.Payload
	case tui.BodyResizeMsg:
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
	return tui.Breadcrumbs("Task", *m.task.Parent)
}

func (m *model) setViewportDimensions(width, height int) {
	m.viewport.Width = max(0, width-2)
	m.viewport.Height = max(0, height-2)
}

// View renders the viewport
func (m model) View() string {
	// The viewport has a fixed width of 80 columns. If there is sufficient
	// additional width in the terminal, then metadata is shown alongside the
	// viewport.
	body := tui.Regular.Copy().
		Margin(0, 1).
		Width(80).
		Render(m.viewport.View())
	if m.width > 104 {
		metadata := tui.Regular.Copy().
			Margin(0, 1).
			Height(m.viewport.Height).
			Width(22).
			Foreground(tui.LightGrey).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			Render(m.task.ID().String())
		// Combine viewport and optional metadata into main body, separated by a
		// vertical rule.
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, metadata)
	}

	// render task footer only if task is not a child of a run model
	if false {
		command := strings.Join(append(m.task.Command, m.task.Args...), " ")
		footer := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("237")).
				Bold(true).
				Padding(0, 1).
				Render(command),
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("36")).
				Width(m.width-tui.Width(command)-2).
				Align(lipgloss.Right).
				Padding(0, 1).
				Render(string(m.task.State)),
		)
		// Combine body and footer.
		body = lipgloss.JoinVertical(lipgloss.Top,
			body,
			strings.Repeat("─", m.width),
			footer,
		)
	}

	return body
}

func (m model) Pagination() string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#a8a7a5")).
		// off white
		Foreground(lipgloss.Color("#FAF9F6")).
		Padding(0, 1).
		Margin(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
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
	msg := outputMsg{taskID: m.task.ID(), isRunTab: m.isRunTab}

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

// TaskCmd returns a command that creates one or more tasks using the given IDs.
func TaskCmd(fn task.Func, ids ...resource.ID) tea.Cmd {
	// Handle the case where a user has pressed a key on an empty table with
	// zero rows
	if len(ids) == 0 {
		return nil
	}

	// If items have been selected then clear the selection
	var deselectCmd tea.Cmd
	if len(ids) > 1 {
		deselectCmd = tui.CmdHandler(table.DeselectMsg{})
	}

	cmd := func() tea.Msg {
		//var task *task.Task
		for _, id := range ids {
			var err error
			if _, err = fn(id); err != nil {
				return tui.NewErrorMsg(err, "creating task")
			}
		}
		//if len(ids) > 1 {
		//	// User has selected multiple rows, so send them to the task *list*
		//	// page
		//	//
		//	// TODO: pass in parameter specifying the parent resource for the
		//	// task listing, i.e. module, workspace, run, etc.
		//	return navigationMsg{
		//		target: page{kind: TaskListKind},
		//	}
		//} else {
		//	// User has highlighted a single row, so send them to the task page.
		//	return navigationMsg{
		//		target: page{kind: TaskKind, resource: task.Resource},
		//	}
		//}
		return nil
	}

	return tea.Batch(cmd, deselectCmd)
}
