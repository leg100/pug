package logs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

type Maker struct {
	Logger *logging.Logger
}

func (mm *Maker) Make(_ resource.Resource, width, height int) (tui.Model, error) {
	columns := []table.Column{
		{Title: "TIME", Width: 24},
		{Title: "LEVEL", Width: 5},
		{Title: "MESSAGE", FlexFactor: 1},
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
	return tui.KeyMapToSlice(viewport.DefaultKeyMap())
}
