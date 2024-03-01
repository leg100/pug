package task

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	taskpkg "github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/common"
)

type listModel struct {
	table table.Model
}

func NewListModel(svc *taskpkg.Service, parent uuid.UUID) listModel {
	tasks := svc.List(taskpkg.ListOptions{
		Ancestor: parent,
	})
	// TODO: construct table widget, determine columns, and populate
	// rows.
	return listModel{table: l}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Modules, Keys.Escape):
			return m, navigate(moduleListState)
		}
	case viewSizeMsg:
		m.table.SetSize(msg.Width, msg.Height)
		return m, nil
	case taskNewMsg:
		task := (*taskpkg.Task)(msg)
		return m, m.table.InsertItem(0, task)
	}

	// Handle keyboard and mouse events in the viewport
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m listModel) Title() string {
	return fmt.Sprintf("tasks (%s)", m.mod)
}

func (m listModel) View() string {
	return m.table.View()
}
