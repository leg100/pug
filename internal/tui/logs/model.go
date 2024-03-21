package logs

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

type Maker struct {
	Logger *logging.Logger
}

func (mm *Maker) Make(_ resource.Resource, width, height int) (tui.Model, error) {
	timeColumn := table.Column{
		Key:   "time",
		Title: "TIME",
		Width: 24,
	}
	levelColumn := table.Column{
		Key:   "level",
		Title: "LEVEL",
		Width: 5,
	}
	msgColumn := table.Column{
		Key:        "message",
		Title:      "MESSAGE",
		FlexFactor: 1,
	}
	columns := []table.Column{
		timeColumn,
		levelColumn,
		msgColumn,
	}
	renderer := func(msg logging.Message, style lipgloss.Style) table.RenderedRow {
		row := table.RenderedRow{
			timeColumn.Key:  msg.Time.Format("2006-01-02T15:04:05.000"),
			levelColumn.Key: msg.Level,
		}

		// combine message and attributes, separated by spaces, with each
		// attribute key/value joined with a '='
		msgAndAttributes := append([]string{msg.Message}, msg.Attributes...)
		row[msgColumn.Key] = strings.Join(msgAndAttributes, " ")
		return row
	}
	table := table.New(columns, renderer, width, height).
		WithSortFunc(logging.ByTimeDesc).
		WithSelectable(false)

	return model{logger: mm.Logger, table: table}, nil
}

type model struct {
	logger *logging.Logger
	table  table.Model[logging.Message]
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[logging.Message](m.logger.Messages)
	}
}

func (m model) Update(msg tea.Msg) (tui.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) Title() string {
	return tui.Bold.Render("Logs")
}

func (m model) View() string {
	return m.table.View()
}

func (m model) Pagination() string {
	return ""
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return nil
}
