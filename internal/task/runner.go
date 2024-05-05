package task

import (
	"context"
	"sync"

	"github.com/leg100/pug/internal/logging"
)

// Runner is the global task Runner that provides a couple of invariants:
// (a) no more than MAX tasks run at any given time
// (b) no more than one 'exclusive' task runs at any given time
type runner struct {
	max   int
	tasks taskLister
}

// StartRunner starts the task runner and returns a function that waits for
// running tasks to finish.
func StartRunner(ctx context.Context, logger logging.Interface, tasks *Service, maxTasks int) func() {
	sub := tasks.Subscribe()
	r := &runner{
		max:   maxTasks,
		tasks: tasks,
	}
	g := sync.WaitGroup{}

	// On each task event, get a list of tasks to be run, start them, and wait
	// for them to complete in the background.
	go func() {
		for range sub {
			for _, task := range r.runnable() {
				waitfn, err := task.start(ctx)
				if err != nil {
					logger.Error("starting task", "error", err.Error(), "task", task)
				} else {
					logger.Debug("started task", "task", task)
					g.Add(1)
					go func() {
						waitfn()
						g.Done()
					}()
				}
			}
		}
	}()

	return g.Wait
}

// runnable retrieves a list of tasks to be run
func (r *runner) runnable() []*Task {
	// exclusive is true if the one and only exclusive slot is occupied
	var exclusive bool

	running := r.tasks.List(ListOptions{
		Status: []Status{Running},
	})
	avail := r.max - len(running)

	// Process queue, starting with oldest task
	queued := r.tasks.List(ListOptions{
		Status: []Status{Queued},
		Oldest: true,
	})
	var i int
	for _, qt := range queued {
		if avail <= 0 && !qt.Immediate {
			// No more available slots. Note: immediate tasks are immediately runnable, so they
			// are exempt from the max. For this reason the number of slots may
			// go into negative territory.
			continue
		}
		if qt.exclusive {
			if exclusive {
				// Exclusive slot taken
				continue
			}
			// Check if there is an exclusive task running
			runningExclusiveTasks := r.tasks.List(ListOptions{
				Exclusive: true,
				Status:    []Status{Running},
			})
			if len(runningExclusiveTasks) > 0 {
				// Exclusive slot taken
				exclusive = true
				continue
			}
			// No exclusive tasks are already running, and this exclusive task
			// takes the available exclusive slot.
			exclusive = true
		}
		avail--
		queued[i] = qt
		i++
	}
	return queued[:i]
}
