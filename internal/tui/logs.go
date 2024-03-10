package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
)

var defaultHighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("#334"))

type logsModelMaker struct {
	logger *logging.Logger
}

func (m *logsModelMaker) makeModel(taskResource resource.Resource) (Model, error) {
	columns := []table.Column{
		table.NewColumn(ColKeyTime, "TIME", 24).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(ColKeyLevel, "LEVEL", 5).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewFlexColumn(ColKeyMessage, "MESSAGE", 30).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}
	rowMaker := func(msg logging.Message) table.RowData {
		data := table.RowData{
			ColKeyTime:  msg.Time.Format("2006-01-02T15:04:05.000"),
			ColKeyLevel: msg.Level,
		}

		// combine message and attributes, separated by spaces, with each
		// attribute key/value joined with a '='
		parts := make([]string, 1, len(msg.Attributes)+1)
		parts[0] = msg.Message
		for k, v := range msg.Attributes {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		data[ColKeyMessage] = strings.Join(parts, " ")
		return data
	}
	return logsModel{
		logger: m.logger,
		table: newTableModel(tableModelOptions[logging.Message]{
			rowMaker:          rowMaker,
			columns:           columns,
			disableSelections: true,
		}),
	}, nil
}

type logMsg string

type logsModel struct {
	logger *logging.Logger
	table  tableModel[logging.Message]
}

func (m logsModel) Init() tea.Cmd {
	return func() tea.Msg {
		return bulkInsertMsg[logging.Message](m.logger.Messages)
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
		Width(m.table.width).
		Padding(0, 1).
		Render("logs")
}

func (m logsModel) View() string {
	return lipgloss.NewStyle().
		Render(m.table.View())
}

func (m logsModel) Pagination() string {
	return m.table.Pagination()
}

func (m logsModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, Keys.CloseHelp)
	return
}
