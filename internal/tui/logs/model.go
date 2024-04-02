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

const timeFormat = "2006-01-02T15:04:05.000"

var (
	timeColumn = table.Column{
		Key:   "time",
		Title: "TIME",
		Width: len(timeFormat),
	}
	levelColumn = table.Column{
		Key:   "level",
		Title: "LEVEL",
		Width: 5, // Width of widest level, ERROR
	}
	msgColumn = table.Column{
		Key:        "message",
		Title:      "MESSAGE",
		FlexFactor: 1,
	}
)

type Maker struct {
	Logger *logging.Logger
}

func (mm *Maker) Make(_ resource.Resource, width, height int) (tui.Model, error) {
	columns := []table.Column{
		timeColumn,
		levelColumn,
		msgColumn,
	}
	renderer := func(msg logging.Message, inherit lipgloss.Style) table.RenderedRow {
		var levelStyle lipgloss.Style
		switch msg.Level {
		case "ERROR":
			levelStyle = tui.Regular.Foreground(tui.Red)
		case "WARN":
			levelStyle = tui.Regular.Foreground(tui.Orange)
		case "DEBUG":
			levelStyle = tui.Regular.Foreground(tui.Grey)
		default:
			levelStyle = tui.Regular.Foreground(tui.Black)
		}
		row := table.RenderedRow{
			timeColumn.Key:  msg.Time.Format(timeFormat),
			levelColumn.Key: levelStyle.Copy().Render(msg.Level),
		}

		// combine message and attributes, separated by spaces, with each
		// attribute key/value joined with a '='
		var b strings.Builder
		b.WriteString(msg.Message)
		b.WriteRune(' ')
		for _, attr := range msg.Attributes {
			b.WriteString(tui.Regular.Copy().Inherit(inherit).Foreground(tui.Black).Render(attr.Key + "="))
			b.WriteString(tui.Regular.Copy().Inherit(inherit).Render(attr.Value + " "))
		}

		row[msgColumn.Key] = lipgloss.NewStyle().Inherit(inherit).Render(b.String())
		return row
	}
	table := table.New[uint](columns, renderer, width, height).
		WithSortFunc(logging.BySerialDesc).
		WithSelectable(false)

	return model{logger: mm.Logger, table: table}, nil
}

type model struct {
	logger *logging.Logger
	table  table.Model[uint, logging.Message]
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

	switch msg := msg.(type) {
	case table.BulkInsertMsg[logging.Message]:
		existing := m.table.Items()
		for _, m := range msg {
			existing[m.Serial] = m
		}
		m.table.SetItems(existing)
	case resource.Event[logging.Message]:
		switch msg.Type {
		case resource.CreatedEvent:
			existing := m.table.Items()
			existing[msg.Payload.Serial] = msg.Payload
			m.table.SetItems(existing)
		}
	}

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