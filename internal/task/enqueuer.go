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

	e := enqueuer{
		tasks: tasks,
	}

	for event := range sub {
		for _, t := range e.Enqueue(event) {
			// TODO: log error
			_, _ = tasks.Enqueue(t.Resource)
		}
	}
}

// Enqueue uses a recent task event to nominate tasks for enqueuing.
// Algorithm:
// * a task can only be enqueued if:
//   - no task belonging to parent resources is active and blocking
func (e *enqueuer) Enqueue(event resource.Event[*Task]) []*Task {
	// Retrieve active tasks
	active := e.tasks.List(ListOptions{
		Status: []Status{Queued, Running},
	})
	// Build a set of active blocked resources
	blocked := make(map[resource.Resource]struct{})
	for _, t := range active {
		if t.Blocking {
			blocked[*t.Parent] = struct{}{}
		}
	}

	// Retrieve module's pending tasks in order of oldest first.
	pending := e.tasks.List(ListOptions{
		Status: []Status{Pending},
		Oldest: true,
	})
	// Remove tasks from pending that should not be enqueued
	for i, t := range pending {
		if _, ok := blocked[*t.Parent]; ok {
			// Remove blocked task
			pending = append(pending[:i], pending[i+1:]...)
		} else if t.Blocking {
			// Found blocking task in pending queue; no further tasks shall be
			// enqueued for the workspace
			blocked[name] = struct{}{}
		}
	}
	// Enqueue filtered pending tasks
	return pending
}
