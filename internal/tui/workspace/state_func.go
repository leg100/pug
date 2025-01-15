package workspace

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
)

type stateFunc func(workspaceID resource.Identity, addr state.ResourceAddress) (task.Spec, error)

func (m resourceList) createStateCommand(fn stateFunc, addrs ...state.ResourceAddress) tea.Cmd {
	// Make N copies of the workspace ID where N is the number of addresses
	workspaceIDs := make([]resource.Identity, len(addrs))
	for i := range workspaceIDs {
		workspaceIDs[i] = m.workspace.ID
	}
	f := newStateTaskFunc(fn, addrs...)
	return m.CreateTasks(f.createTask, workspaceIDs...)
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

func (f *stateTaskFunc) createTask(workspaceID resource.Identity) (task.Spec, error) {
	t, err := f.fn(workspaceID, f.addrs[f.i])
	f.i++
	return t, err
}
