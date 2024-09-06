package task

import (
	"context"

	"github.com/leg100/pug/internal/resource"
)

// enqueuer determines which tasks shall be added to the global queue. A task
// can be enqueued if:
//
// (a) it is an "immediate" task, or:
// (b) if it belongs to a workspace then no other task has "blocked" that workspace
// (c) if it belongs to a module then no other task has "blocked" that module
// (d) if it has dependencies on other tasks then those tasks have all finished
// successfully.
//
// Otherwise the enqueuer leaves the task in a pending state.
type enqueuer struct {
	tasks enqueuerTaskService
}

type enqueuerTaskService interface {
	taskLister

	Get(taskID resource.ID) (*Task, error)
}

func StartEnqueuer(tasks *Service) {
	e := enqueuer{tasks: tasks}
	sub := tasks.TaskBroker.Subscribe(context.Background())

	go func() {
		for range sub {
			for _, t := range e.enqueuable() {
				tasks.Enqueue(t.ID)
			}
		}
	}()
}

// enqueuable returns a list of a tasks to be moved from the pending state to the
// queued state.
func (e *enqueuer) enqueuable() []*Task {
	// Retrieve active tasks
	active := e.tasks.List(ListOptions{
		Status: []Status{Queued, Running},
	})
	// blockedModules are those modules blocked by tasks: the keys are the IDs
	// of the modules and the values are the IDs of tasks blocking the
	// respective module.
	blockedModules := make(map[resource.ID]struct{})
	// blockedWorkspaces are those workspaces blocked by tasks: the keys are the IDs
	// of the workspaces and the values are the IDs of tasks blocking the
	// respective workspace.
	blockedWorkspaces := make(map[resource.ID]struct{})
	// Populate set of currently blocked workspaces/modules.
	for _, t := range active {
		if t.Blocking {
			if t.ModuleID != nil {
				blockedModules[*t.ModuleID] = struct{}{}
			}
			if t.WorkspaceID != nil {
				blockedWorkspaces[*t.WorkspaceID] = struct{}{}
			}
		}
	}
	// Retrieve pending tasks in order of oldest first.
	pending := e.tasks.List(ListOptions{
		Status: []Status{Pending},
		Oldest: true,
	})
	// Build list of tasks to enqueue
	var enqueue []*Task
	for _, t := range pending {
		if t.Immediate {
			// Always enqueue immediate tasks.
			enqueue = append(enqueue, t)
			continue
		}
		if t.WorkspaceID != nil {
			if _, ok := blockedWorkspaces[*t.WorkspaceID]; ok {
				// Don't enqueue task belonging to workspace blocked by another task
				continue
			}
		}
		if t.ModuleID != nil {
			if _, ok := blockedModules[*t.ModuleID]; ok {
				// Don't enqueue task belonging to module blocked by another task
				continue
			}
		}
		if !e.enqueueDependentTask(t) {
			// Don't enqueue task with dependencies on other tasks that have yet
			// to complete or have failed.
			continue
		}
		// Enqueue task.
		enqueue = append(enqueue, t)
		// Blocking tasks can block workspaces and modules.
		if t.Blocking {
			if t.WorkspaceID != nil {
				// Task blocks workspace; no further tasks belonging to workspace
				// shall be enqueued.
				blockedWorkspaces[*t.WorkspaceID] = struct{}{}
			}
			if t.ModuleID != nil {
				// Task blocks module; no further tasks belonging to module
				// shall be enqueued.
				blockedModules[*t.ModuleID] = struct{}{}
			}
		}
	}
	return enqueue
}

func (e *enqueuer) enqueueDependentTask(t *Task) bool {
	for _, id := range t.DependsOn {
		dependency, err := e.tasks.Get(id)
		if err != nil {
			// TODO: decide what to do in case of error
			return false
		}
		switch dependency.State {
		case Exited:
			// Is enqueuable if all dependencies have exited successfully.
		case Canceled, Errored:
			// Dependency failed so mark task as failed too by cancelling it
			// along with a reason why it was canceled.
			t.stdout.Write([]byte("task dependency failed"))
			t.updateState(Canceled)
			return false
		default:
			// Not enqueueable
			return false
		}
	}
	return true
}
