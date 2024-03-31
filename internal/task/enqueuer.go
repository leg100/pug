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
		for _, t := range e.enqueue() {
			// TODO: log error
			_, _ = tasks.Enqueue(t.ID)
		}
	}
}

// enqueue returns a list of a tasks to be moved from the pending state to the
// queued state.
func (e *enqueuer) enqueue() []*Task {
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
	// Remove tasks from pending that should not be enqueued
	for i, t := range pending {
		// Recursively walk task's ancestors and check if they are currently
		// blocked; if so then task cannot be enqueued.
		if hasBlockedAncestor(blocked, *t.Parent) {
			// Remove from pending
			pending = append(pending[:i], pending[i+1:]...)
		} else if t.Blocking {
			// Found blocking task in pending queue; no further tasks shall be
			// enqueued for resources belonging to the task's parent resource
			blocked[t.Parent.ID] = struct{}{}
		}
	}
	// Enqueue filtered pending tasks
	return pending
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
