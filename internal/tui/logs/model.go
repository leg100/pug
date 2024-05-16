package logs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

var (
	keyColumn = table.Column{
		Key:   "key",
		Title: "KEY",
		Width: 30,
	}
	valueColumn = table.Column{
		Key:        "value",
		Title:      "VALUE",
		FlexFactor: 1,
	}
)

const (
	timeAttrKey    = "time"
	levelAttrKey   = "level"
	messageAttrKey = "message"
)

type Maker struct {
	Logger *logging.Logger
}

func (mm *Maker) Make(r resource.Resource, width, height int) (tea.Model, error) {
	msg, err := mm.Logger.Get(r.GetID())
	if err != nil {
		return model{}, err
	}
	columns := []table.Column{keyColumn, valueColumn}
	renderer := func(attr logging.Attr) table.RenderedRow {
		return table.RenderedRow{
			keyColumn.Key:   attr.Key,
			valueColumn.Key: attr.Value,
		}
	}
	items := map[string]logging.Attr{
		timeAttrKey: {
			Key:   timeAttrKey,
			Value: msg.Time.Format(timeFormat),
		},
		messageAttrKey: {
			Key:   messageAttrKey,
			Value: msg.Message,
		},
		levelAttrKey: {
			Key:   levelAttrKey,
			Value: coloredLogLevel(msg.Level),
		},
	}
	for _, attr := range msg.Attributes {
		items[attr.Key] = attr
	}
	table := table.New[string](columns, renderer, width, height).
		WithSortFunc(byAttribute)
	table.SetItems(items)

	return model{
		msg:    msg,
		table:  table,
		width:  width,
		height: height,
	}, nil
}

type model struct {
	msg    logging.Message
	table  table.Model[string, logging.Attr]
	width  int
	height int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle keyboard and mouse events in the table widget
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) Title() string {
	title := tui.TitleStyle.Render("LogMessage")
	id := tui.Regular.Copy().Foreground(tui.Pink).Render(fmt.Sprintf("#%d", m.msg.Serial))
	s := fmt.Sprintf("%s(%s)", title, id)
	return tui.Bold.Render(s)
}

func (m model) View() string {
	return m.table.View()
}

func (m model) HelpBindings() (bindings []key.Binding) {
	return nil
}

// byAttribute sorts the attributes of an individual message for display in the
// logs model.
func byAttribute(i, j logging.Attr) int {
	switch i.Key {
	case timeAttrKey:
		// time comes first
		return -1
	case levelAttrKey:
		switch j.Key {
		case timeAttrKey:
			return 1
		}
		// then level
		return -1
	case messageAttrKey:
		switch j.Key {
		case timeAttrKey, levelAttrKey:
			return 1
		}
		// then message
		return -1
	}
	// then everything else, in any order
	return 1
}
