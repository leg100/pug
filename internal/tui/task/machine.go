package task

import (
	"bufio"

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
	resourceStatusColumn = table.Column{
		Key:   "resource_status",
		Title: "STATUS",
		Width: 7,
	}
	resourceAddressColumn = table.Column{
		Key:        "resource_address",
		Title:      "ADDRESS",
		FlexFactor: 1,
	}
)

type machineModel struct {
	id      uuid.UUID
	table   table.Model[machineResource]
	scanner *bufio.Scanner
}

type machineModelOptions struct {
	width  int
	height int
}

type machineResource struct {
	Address state.ResourceAddress
	Status  string
}

func (r machineResource) GetID() resource.ID {
	return r.Address
}

func newMachineModel(t *task.Task, opts machineModelOptions) machineModel {
	columns := []table.Column{
		resourceStatusColumn,
		resourceAddressColumn,
	}
	renderer := func(r machineResource) table.RenderedRow {
		return table.RenderedRow{
			resourceStatusColumn.Key:  string(r.Status),
			resourceAddressColumn.Key: string(r.Address),
		}
	}
	tbl := table.New(
		columns,
		renderer,
		opts.width,
		opts.height,
		table.WithSortFunc(func(i, j machineResource) int {
			if i.Address < j.Address {
				return -1
			} else {
				return 1
			}
		}),
	)
	return machineModel{
		id:      uuid.New(),
		table:   tbl,
		scanner: bufio.NewScanner(t.NewReader(false)),
	}
}

func (m machineModel) Init() tea.Cmd {
	return m.getOutput
}

func (m machineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case machineOutputMsg:
		// Ensure output is for this model
		if msg.modelID != m.id {
			return m, nil
		}
		if msg.eof {
			return m, nil
		}
		mm, err := machine.UnmarshalMessage(msg.line)
		if err != nil {
			return m, tui.ReportError(err)
		}
		switch mm := mm.(type) {
		case *machine.PlannedChangeMsg:
			mr := machineResource{
				Address: state.ResourceAddress(mm.Change.Resource.ResourceName),
				Status:  string(mm.Type),
			}
			m.table.AddItems(mr)
		}
	}
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m machineModel) View() string {
	return m.table.View()
}

func (m machineModel) getOutput() tea.Msg {
	msg := machineOutputMsg{modelID: m.id}
	if m.scanner.Scan() {
		msg.line = m.scanner.Bytes()
	} else {
		msg.eof = true
	}
	return msg
}

type machineOutputMsg struct {
	modelID uuid.UUID
	line    []byte
	eof     bool
}
