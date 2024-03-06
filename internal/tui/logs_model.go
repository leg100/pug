package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/common"
)

type logsModelMaker struct {
	logger *logging.Logger
}

func (m *logsModelMaker) makeModel(taskResource resource.Resource) (common.Model, error) {
	columns := []table.Column{
		table.NewColumn(common.ColKeyTime, "TIME", 20).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewColumn(common.ColKeyLevel, "LEVEL", 10).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
		table.NewFlexColumn(common.ColKeyMessage, "MESSAGE", 30).WithStyle(
			lipgloss.NewStyle().
				Align(lipgloss.Left),
		),
	}
	return logsModel{
		logger: m.logger,
		table:  table.New(columns).Focused(true).WithMultiline(true),
	}, nil
}

type logMsg string

type logsModel struct {
	logger   *logging.Logger
	table    table.Model
	messages []logging.Message
}

func (m logsModel) Init() tea.Cmd {
	return func() tea.Msg {
		return common.BulkInsertMsg[logging.Message](m.logger.Messages)
	}
}

func (m logsModel) Update(msg tea.Msg) (common.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case common.BulkInsertMsg[logging.Message]:
		m.messages = msg
		m.table = m.table.WithRows(m.toRows())
	case resource.Event[logging.Message]:
		m.messages = append(m.messages, msg.Payload)
		m.table = m.table.WithRows(m.toRows())
	case common.ViewSizeMsg:
		// Accomodate margin of size 1 on either side
		m.table = m.table.WithTargetWidth(msg.Width - 2)
		//m.table = m.table.WithMinimumHeight(msg.Height)
		return m, nil
	}
	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m logsModel) Title() string {
	return "logs"
}

func (m logsModel) View() string {
	return lipgloss.NewStyle().
		Margin(0, 1).
		Render(m.table.View())
}

func (m logsModel) HelpBindings() (bindings []key.Binding) {
	bindings = append(bindings, common.Keys.CloseHelp)
	return
}

func (m logsModel) toRows() []table.Row {
	rows := make([]table.Row, len(m.messages))
	for i, msg := range m.messages {
		data := table.RowData{
			common.ColKeyTime:  msg.Time.Round(time.Second),
			common.ColKeyLevel: msg.Level,
		}
		switch msg.Level {
		case "ERROR":
			data[common.ColKeyMessage] = fmt.Sprintf("%s: %s", msg.Message, msg.Error)
		default:
			data[common.ColKeyMessage] = msg.Message
		}
		rows[i] = table.NewRow(data)
	}
	return rows
}
