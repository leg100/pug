package task

import (
	"context"

	"github.com/leg100/pug/internal/resource"
)

// enqueuer determines which tasks should be added to the global queue for
// processing
type enqueuer struct {
	tasks taskLister
}

func StartEnqueuer(ctx context.Context, tasks *Service) {
	e := enqueuer{tasks: tasks}
	sub := tasks.Broker.Subscribe(ctx)

	for range sub {
		for _, t := range e.enqueuable() {
			tasks.Enqueue(t.ID)
		}
	}
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
			blocked[ab.Parent.ID] = struct{}{}
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
		if !t.Immediate && hasBlockedAncestor(blocked, *t.Parent) {
			// Not enqueuable
			continue
		} else if t.Blocking {
			// Found blocking task in pending queue; no further tasks shall be
			// enqueued for resources belonging to the task's parent resource
			blocked[t.Parent.ID] = struct{}{}
		}
		// Enqueueable
		pending[i] = t
		i++
	}
	// Enqueue filtered pending tasks
	return pending[:i]
}

func hasBlockedAncestor(blocked map[resource.ID]struct{}, parent resource.Resource) bool {
	if _, ok := blocked[parent.ID]; ok {
		return true
	}
	if parent.Parent != nil {
		return hasBlockedAncestor(blocked, *parent.Parent)
	}
	return false
}
