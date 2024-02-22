package task

import (
	"context"
	"log/slog"

	"github.com/leg100/pug/internal/pubsub"
)

// Runner is the global task Runner that provides a couple of invariants:
// (a) no more than MAX tasks run at any given time
// (b) no more than one 'exclusive' task runs at any given time
// (c)
type runner struct {
	max       int
	exclusive chan struct{}

	broker *pubsub.Broker[*Task]
	tasks  taskLister
}

type taskLister interface {
	List(opts ListOptions) []*Task
}

func newRunner(maxTasks int, lister taskLister) *runner {
	return &runner{
		max:       maxTasks,
		exclusive: make(chan struct{}, 1),
		broker:    pubsub.NewBroker[*Task](),
		tasks:     lister,
	}
}

func (r *runner) start(ctx context.Context) {
	sub, unsub := r.broker.Subscribe(ctx)
	defer unsub()

	// Each new event triggers the runner to attempt to run tasks
	for range sub {
		r.process()
	}
}

func (r *runner) process() {
	running := r.tasks.List(ListOptions{
		Status: []Status{Running},
	})
	avail := r.max - len(running)
	if avail == 0 {
		// Already running maximum number
		return
	}
	// Process queue, starting with oldest task
	queued := r.tasks.List(ListOptions{
		Status: []Status{Queued},
		Oldest: true,
	})
	for _, qt := range queued {
		if avail == 0 {
			break
		}
		if qt.exclusive {
			// Only run task if exclusive slot is empty
			select {
			case r.exclusive <- struct{}{}:
				r.run(qt)
				avail--
			default:
			}
		} else {
			r.run(qt)
			avail--
		}
	}
}

func (r *runner) run(t *Task) {
	waitfn, err := t.start()
	if err != nil {
		slog.Error(err.Error())
	}
	go waitfn()
}
