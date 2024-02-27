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

func startEnqueuer(
	ctx context.Context,
	tasks *Service,
) {

	sub, unsub := tasks.Watch(ctx)
	defer unsub()

	e := enqueuer{tasks: tasks}

	for range sub {
		for _, t := range e.Enqueue() {
			// TODO: log error
			_, _ = tasks.Enqueue(t.Resource)
		}
	}
}

// Enqueue returns a list of a tasks to be moved from the pending state to the
// queued state.
func (e *enqueuer) Enqueue() []*Task {
	// Retrieve active tasks
	active := e.tasks.List(ListOptions{
		Status: []Status{Queued, Running},
	})
	// Construct set of task owners that are currently blocked.
	blocked := make(map[resource.Resource]struct{}, len(active))
	for _, ab := range active {
		if ab.Blocking {
			blocked[*ab.Parent] = struct{}{}
		}
	}

	// Retrieve pending tasks in order of oldest first.
	pending := e.tasks.List(ListOptions{
		Status: []Status{Pending},
		Oldest: true,
	})
	// Remove tasks from pending that should not be enqueued
	for i, t := range pending {
		// Recursively walk task's owners and check if they are currently
		// blocked; if so then task cannot be enqueued.
		if hasBlockedParent(blocked, *t.Parent) {
			pending = append(pending[:i], pending[i+1:]...)
		} else if t.Blocking {
			// Found blocking task in pending queue; no further tasks shall be
			// enqueued for resources belonging to the task's parent resource
			blocked[*t.Parent] = struct{}{}
		}
	}
	// Enqueue filtered pending tasks
	return pending
}

func hasBlockedParent(blocked map[resource.Resource]struct{}, r resource.Resource) bool {
	if _, ok := blocked[r]; ok {
		return true
	}
	if r.Parent != nil {
		return hasBlockedParent(blocked, *r.Parent)
	}
	return false
}
