package task

import (
	"context"

	"github.com/leg100/pug/internal/pubsub"
)

// Runner is the global task Runner that provides a couple of invariants:
// (a) no more than MAX tasks run at any given time
// (b) no more than one 'exclusive' task runs at any given time
// (c)
type scheduler struct {
	max int

	broker *pubsub.Broker[*Task]
	cache  *Categories
}

func newScheduler(maxTasks int) *scheduler {
	return &scheduler{
		max:    maxTasks,
		broker: pubsub.NewBroker[*Task](),
		cache:  &Categories{},
	}
}

func (r *scheduler) start(ctx context.Context) {
	// Only start task once there is an available slot, and upon completion
	// unoccupy the slot; if this is an exclusive task then use the
	// exclusive slot.
	sub, unsub := r.broker.Subscribe(ctx)
	defer unsub()

	for event := range sub {
		r.cache.Categorize(event)
		enqueue := event.Payload.Scheduler.Handle(event)
		for _, t := range enqueue {
			t.updateState(Queued)
		}
		// Only transition task from queued to running once there is an
		// available slot.
		if len(r.cache.Running) < r.maxTasks {
		}
	}
	if exclusive {
		r.exclusive <- struct{}{}
		defer func() { <-r.exclusive }()
	} else {
		r.tasks <- struct{}{}
		defer func() { <-r.tasks }()
	}
	// start and wait for task to finish
	if waitfunc := task.run(); waitfunc != nil {
		waitfunc()
	}
}
