package logs

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
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
		Width: len("ERROR"),
	}
	msgColumn = table.Column{
		Key:        "message",
		Title:      "MESSAGE",
		FlexFactor: 1,
	}
)

type ListMaker struct {
	Logger *logging.Logger
}

func (m *ListMaker) Make(_ resource.Resource, width, height int) (tea.Model, error) {
	columns := []table.Column{
		timeColumn,
		levelColumn,
		msgColumn,
	}
	renderer := func(msg logging.Message) table.RenderedRow {
		// combine message and attributes, separated by spaces, with each
		// attribute key/value joined with a '='
		var b strings.Builder
		b.WriteString(msg.Message)
		b.WriteRune(' ')
		for _, attr := range msg.Attributes {
			b.WriteString(tui.Regular.Copy().Faint(true).Render(attr.Key + "="))
			b.WriteString(tui.Regular.Copy().Render(attr.Value + " "))
		}

		return table.RenderedRow{
			timeColumn.Key:  msg.Time.Format(timeFormat),
			levelColumn.Key: coloredLogLevel(msg.Level),
			msgColumn.Key:   tui.Regular.Copy().Render(b.String()),
		}
	}
	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(logging.BySerialDesc),
		table.WithSelectable[logging.Message](false),
	)

	return list{logger: m.Logger, table: table}, nil
}

type list struct {
	logger *logging.Logger
	table  table.Model[logging.Message]
}

func (m list) Init() tea.Cmd {
	return func() tea.Msg {
		return table.BulkInsertMsg[logging.Message](m.logger.List())
	}
}

func (m list) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.table.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.LogKind, tui.WithParent(row.Value))
			}
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m list) Title() string {
	return tui.Breadcrumbs("Logs", resource.GlobalResource)
}

func (m list) View() string {
	return m.table.View()
}

func (m list) HelpBindings() (bindings []key.Binding) {
	return nil
}
