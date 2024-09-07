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
	Logger  *logging.Logger
	Helpers *tui.Helpers
}

func (mm *Maker) Make(id resource.ID, width, height int) (tea.Model, error) {
	msg, err := mm.Logger.Get(id)
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
	table := table.New(columns, renderer, width, height,
		table.WithSortFunc(byAttribute),
		table.WithSelectable[logging.Attr](false),
	)
	items := []logging.Attr{
		{
			Key:   timeAttrKey,
			Value: msg.Time.Format(timeFormat),
			ID:    resource.NewID(resource.LogAttr),
		},
		{
			Key:   messageAttrKey,
			Value: msg.Message,
			ID:    resource.NewID(resource.LogAttr),
		},
		{
			Key:   levelAttrKey,
			Value: coloredLogLevel(msg.Level),
			ID:    resource.NewID(resource.LogAttr),
		},
	}
	items = append(items, msg.Attributes...)
	table.SetItems(items...)

	return model{
		msg:    msg,
		table:  table,
		width:  width,
		height: height,
	}, nil
}

type model struct {
	msg    logging.Message
	table  table.Model[logging.Attr]
	width  int
	height int

	*tui.Helpers
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) Title() string {
	serial := tui.TitleSerial.Render(fmt.Sprintf("#%d", m.msg.Serial))
	return m.Breadcrumbs("LogMessage", nil, serial)
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
