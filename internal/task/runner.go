package task

import (
	"context"
	"log/slog"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
)

// Runner is the global task Runner that provides a couple of invariants:
// (a) no more than MAX tasks run at any given time
// (b) no more than one 'exclusive' task runs at any given time
type runner struct {
	max       int
	exclusive chan struct{}

	broker *pubsub.Broker[*Task]
	tasks  taskLister
}

func newRunner(maxTasks int, lister taskLister) *runner {
	return &runner{
		max:       maxTasks,
		exclusive: make(chan struct{}, 1),
		broker:    pubsub.NewBroker[*Task](),
		tasks:     lister,
	}
}

func (r *runner) start(ctx context.Context, sub <-chan resource.Event[*Task]) {
	//g, ctx := errgroup.WithContext(ctx)
	// On each task event, get a list of tasks to be run, start them, and wait
	// for them to complete in the background.
	for range sub {
		for _, task := range r.runnable() {
			waitfn, err := task.start()
			if err != nil {
				slog.Error(err.Error())
			} else {
				go waitfn()
			}
		}
	}
}

// runnable retrieves a list of tasks to be run
func (r *runner) runnable() []*Task {
	running := r.tasks.List(ListOptions{
		Status: []Status{Running},
	})
	avail := r.max - len(running)
	if avail == 0 {
		// Hit max, can't run any more tasks
		return nil
	}
	// Process queue, starting with oldest task
	queued := r.tasks.List(ListOptions{
		Status: []Status{Queued},
		Oldest: true,
	})
	for i, qt := range queued {
		if avail == 0 {
			return queued[:i]
		}
		if qt.exclusive {
			// Only run task if exclusive slot is empty
			select {
			case r.exclusive <- struct{}{}:
				avail--
			default:
				// Exclusive slot is full; skip task
				queued = append(queued[:i], queued[i+1:]...)
			}
		} else {
			avail--
		}
	}
	return queued
}
