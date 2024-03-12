package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/table"
)

type logsModelMaker struct {
	logger *logging.Logger
}

func (mm *logsModelMaker) makeModel(taskResource resource.Resource) (Model, error) {
	columns := []table.Column{
		{Title: "TIME", Width: 24},
		{Title: "LEVEL", Width: 5},
		// make flex
		{Title: "MESSAGE", Width: 60, FlexFactor: 1},
	}
	cellsFunc := func(msg logging.Message) []table.Cell {
		cells := []table.Cell{
			{Str: msg.Time.Format("2006-01-02T15:04:05.000")},
			{Str: msg.Level},
		}

		// combine message and attributes, separated by spaces, with each
		// attribute key/value joined with a '='
		parts := make([]string, 1, len(msg.Attributes)+1)
		parts[0] = msg.Message
		for k, v := range msg.Attributes {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		return append(cells, table.Cell{Str: strings.Join(parts, " ")})
	}
	table := table.New[logging.Message](columns).
		WithCellsFunc(cellsFunc).
		WithSortFunc(logging.ByTimeDesc).
		WithSelectable(false)

	return logsModel{logger: mm.logger, table: table}, nil
}

type logsModel struct {
	logger *logging.Logger
	table  table.Model[logging.Message]
}

func (m logsModel) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[logging.Message](m.logger.Messages)
	}
}

func (m logsModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m logsModel) Title() string {
	return lipgloss.NewStyle().
		Background(DarkGrey).
		Foreground(White).
		Bold(true).
		//Width(m.table.width).
		Padding(0, 1).
		Render("logs")
}

func (m logsModel) View() string {
	return lipgloss.NewStyle().
		Render(m.table.View())
}

func (m logsModel) Pagination() string {
	return ""
}

func (m logsModel) HelpBindings() (bindings []key.Binding) {
	return keyMapToSlice(m.table.KeyMap)
}
