package task

import (
	"bufio"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/machine"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
)

var (
	resourceActionColumn = table.Column{
		Key:   "resource_action",
		Title: "ACTION",
		Width: 7,
	}
	resourceStatusColumn = table.Column{
		Key:        "resource_status",
		Title:      "STATUS",
		FlexFactor: 1,
	}
	resourceAddressColumn = table.Column{
		Key:        "resource_address",
		Title:      "ADDRESS",
		FlexFactor: 1,
	}
	resourceChangeDuration = table.Column{
		Key:        "time_taken",
		Title:      "TIME",
		FlexFactor: 1,
	}
)

type structuredModel struct {
	id         uuid.UUID
	table      table.Model[structuredResource]
	scanner    *bufio.Scanner
	startTimes map[state.ResourceAddress]time.Time
}

type structuredModelOptions struct {
	width  int
	height int
}

type structuredResource struct {
	Action    string
	Address   state.ResourceAddress
	Status    string
	timeTaken time.Duration
}

func (r structuredResource) GetID() resource.ID {
	return r.Address
}

func newStructuredModel(t *task.Task, opts structuredModelOptions) structuredModel {
	columns := []table.Column{
		resourceActionColumn,
		resourceStatusColumn,
		resourceAddressColumn,
		resourceChangeDuration,
	}
	renderer := func(r structuredResource) table.RenderedRow {
		return table.RenderedRow{
			resourceActionColumn.Key:   string(r.Action),
			resourceStatusColumn.Key:   string(r.Status),
			resourceAddressColumn.Key:  string(r.Address),
			resourceChangeDuration.Key: fmt.Sprintf("%.fs", r.timeTaken.Seconds()),
		}
	}
	tbl := table.New(
		columns,
		renderer,
		opts.width,
		opts.height,
		table.WithSortFunc(func(i, j structuredResource) int {
			if i.Address < j.Address {
				return -1
			} else {
				return 1
			}
		}),
	)
	return structuredModel{
		id:         uuid.New(),
		table:      tbl,
		scanner:    bufio.NewScanner(t.NewReader(false)),
		startTimes: make(map[state.ResourceAddress]time.Time),
	}
}

func (m structuredModel) Init() tea.Cmd {
	return m.getOutput
}

func (m structuredModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case outputMsg:
		// Ensure output is for this model
		if msg.modelID != m.id {
			return m, nil
		}
		if msg.eof {
			return m, nil
		}
		mm, err := machine.UnmarshalMessage(msg.output)
		if err != nil {
			return m, tui.ReportError(err)
		}
		switch mm := mm.(type) {
		case *machine.PlannedChangeMsg:
			mr := structuredResource{
				Action:  string(mm.Change.Action),
				Address: state.ResourceAddress(mm.Change.Resource.Addr),
				Status:  string(mm.Type),
			}
			addr := state.ResourceAddress(mm.Change.Resource.Addr)
			started, ok := m.startTimes[addr]
			if !ok {
				started = mm.TimeStamp
				m.startTimes[addr] = mm.TimeStamp
			}
			mr.timeTaken = mm.TimeStamp.Sub(started)
			m.table.AddItems(mr)
		}
		return m, m.getOutput
	}
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m structuredModel) View() string {
	return m.table.View()
}

func (m structuredModel) getOutput() tea.Msg {
	msg := outputMsg{modelID: m.id}
	if m.scanner.Scan() {
		msg.output = m.scanner.Bytes()
	} else {
		msg.eof = true
	}
	return msg
}
