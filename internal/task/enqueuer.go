package task

import (
	"github.com/leg100/pug/internal/resource"
)

// enqueuer determines which tasks should be added to the global queue for
// processing
type enqueuer struct {
	tasks taskLister
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
	// Construct set of task owners that are currently blocked.
	blocked := make(map[resource.ID]struct{}, len(active))
	for _, ab := range active {
		if ab.Blocking {
			blocked[ab.Parent.GetID()] = struct{}{}
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
		// Recursively walk task's ancestors and check if they are currently
		// blocked; if so then task cannot be enqueued. The exception to this
		// rule is an immediate task, which is always enqueuable
		if !t.Immediate && hasBlockedAncestor(blocked, t.GetParent()) {
			// Not enqueuable
			continue
		} else if t.Blocking {
			// Found blocking task in pending queue; no further tasks shall be
			// enqueued for resources belonging to the task's parent resource
			blocked[t.Parent.GetID()] = struct{}{}
		}
		// Check if task depends upon successful completion of
		// other tasks.
		for _, dep := range t.DependsOn {
			switch dep.State {
			case Exited:
				// Is enqueuable if all dependencies have exited successfully.
			case Canceled, Errored:
				// Dependency failed so mark task as failed too.
				// Write message to task output first to tell user why it failed
				t.buf.Write([]byte("task dependency failed"))
				t.updateState(Canceled)
				continue
			default:
				// Not enqueueable
				continue
			}
		}
		// Enqueueable
		pending[i] = t
		i++
	}
	// Enqueue filtered pending tasks
	return pending[:i]
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
