package task

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/task"
)

type machineModelView int

const (
	structuredView machineModelView = iota
	rawView
)

type machineModel struct {
	views   map[machineModelView]tea.Model
	current machineModelView
}

type machineModelOptions struct {
	disableAutoscroll bool
	spinner           *spinner.Model
	width             int
	height            int
}

func newMachineModel(t *task.Task, opts machineModelOptions) machineModel {
	views := map[machineModelView]tea.Model{
		structuredView: newStructuredModel(t, structuredModelOptions{
			width:  opts.width,
			height: opts.height,
		}),
		rawView: newRaw(t, rawOptions(opts)),
	}
	return machineModel{
		views:   views,
		current: rawView,
	}
}

func (m machineModel) Init() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.views))
	for i, v := range m.views {
		cmds[i] = v.Init()
	}
	return tea.Batch(cmds...)
}

func (m machineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, localKeys.Raw):
			m.current = rawView
		case key.Matches(msg, localKeys.Structured):
			m.current = structuredView
		}
	}

	cmds := make([]tea.Cmd, len(m.views))
	var i int
	for k, v := range m.views {
		v, cmds[i] = v.Update(msg)
		m.views[k] = v
		i++
	}
	return m, tea.Batch(cmds...)
}

func (m machineModel) View() string {
	return m.views[m.current].View()
}
