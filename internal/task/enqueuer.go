package task

import (
	"github.com/leg100/pug/internal/resource"
)

// enqueuer determines which tasks should be added to the global queue for
// processing
type enqueuer struct {
	tasks enqueuerTaskService
}

type enqueuerTaskService interface {
	taskLister

	Get(taskID resource.ID) (*Task, error)
}

func StartEnqueuer(tasks *Service) {
	e := enqueuer{tasks: tasks}
	sub := tasks.TaskBroker.Subscribe()

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
	// Populate set of currently blocked workspaces/modules.
	blocked := make(map[resource.ID]struct{}, len(active))
	for _, t := range active {
		if t.Blocking {
			addBlockedResource(blocked, t.Parent)
		}
	}
	// Retrieve pending tasks in order of oldest first.
	pending := e.tasks.List(ListOptions{
		Status: []Status{Pending},
		Oldest: true,
	})
	// Filter pending tasks, keeping only those that are enqueuable
	var i int
	for _, t := range pending {
		// Check whether task's workspace/module is currently blocked; if so
		// then task cannot be enqueued. The exception to this rule is an
		// immediate task, which is always enqueuable
		if !t.Immediate && hasBlockedAncestor(blocked, t.GetParent()) {
			// Not enqueuable
			continue
		} else if t.Blocking {
			// Found blocking task in pending queue; no further tasks shall be
			// enqueued for resources belonging to any of the blocking task's
			// ancestor resources
			addBlockedResource(blocked, t.Parent)
		}
		// If task depends upon other tasks then only enqueue task if they have
		// all successfully completed.
		if !e.enqueueDependentTask(t) {
			continue
		}
		// Enqueueable
		pending[i] = t
		i++
	}
	// Enqueue filtered pending tasks
	return pending[:i]
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

func addBlockedResource(blocked map[resource.ID]struct{}, parent resource.Resource) {
	if parent != nil {
		switch parent.GetKind() {
		case resource.Module, resource.Workspace:
			blocked[parent.GetID()] = struct{}{}
			return
		}
	}
	if parent.GetParent() != nil {
		addBlockedResource(blocked, parent.GetParent())
	}
}

func hasBlockedAncestor(blocked map[resource.ID]struct{}, parent resource.Resource) bool {
	if _, ok := blocked[parent.GetID()]; ok {
		return true
	}
	if parent.GetParent() != nil {
		return hasBlockedAncestor(blocked, parent.GetParent())
	}
	return false
}
