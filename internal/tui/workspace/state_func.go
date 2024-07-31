package workspace

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
)

type stateFunc func(workspaceID resource.ID, addr state.ResourceAddress) (task.Spec, error)

func (m resourceList) createStateCommand(fn stateFunc, addrs ...state.ResourceAddress) tea.Cmd {
	// Make N copies of the workspace ID where N is the number of addresses
	workspaceIDs := make([]resource.ID, len(addrs))
	for i := range workspaceIDs {
		workspaceIDs[i] = m.workspace.GetID()
	}
	f := newStateTaskFunc(fn, addrs...)
	return m.helpers.CreateTasks(f.createTask, workspaceIDs...)
}

func newStateTaskFunc(fn stateFunc, addrs ...state.ResourceAddress) *stateTaskFunc {
	return &stateTaskFunc{
		fn:    fn,
		addrs: addrs,
	}
}

type stateTaskFunc struct {
	fn    stateFunc
	addrs []state.ResourceAddress
	i     int
}

func (f *stateTaskFunc) createTask(workspaceID resource.ID) (task.Spec, error) {
	t, err := f.fn(workspaceID, f.addrs[f.i])
	f.i++
	return t, err
}
