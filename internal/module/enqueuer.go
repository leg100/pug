package module

import (
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

// Enqueuer schedules tasks
type Enqueuer struct {
	modules *Service
	tasks   *task.Service
}

func (s *Enqueuer) Enqueue(event resource.Event[*task.Task]) []*task.Task {
	// Only schedule tasks if module is initialized
	if mod, err := s.modules.Get(event.Payload.Path); err != nil {
		// log error
		return nil
	} else if mod.Status != Initialized {
		return nil
	}

	// Retrieve module's active tasks
	active := s.tasks.List(task.ListOptions{
		Path:   &event.Payload.Path,
		Status: []task.Status{task.Queued, task.Running},
	})
	if len(active) > 0 && active[0].Kind == InitTask {
		// An active InitTask blocks pending tasks
		return nil
	}

	// Retrieve module's pending tasks in order of oldest first.
	pending := s.tasks.List(task.ListOptions{
		Path:   &event.Payload.Path,
		Status: []task.Status{task.Pending},
		Oldest: true,
	})
	if len(pending) == 0 {
		// Nothing to enqueue.
		return nil
	}
	// If the oldest pending task is an init task then enqueue only that task.
	// Otherwise enqueue as many of the oldest pending tasks until the oldest
	// init task is encountered.
	init := -1
	for i := range pending {
		if pending[i].Kind == InitTask {
			init = i
			break
		}
	}
	switch init {
	case 0:
		// oldest pending task is an init task so just enqueue that
		return pending[0:1]
	case -1:
		// no init task found, so enqueue all
		return pending
	default:
		// init task found, so enqueue all tasks before it
		return pending[0:init]
	}
}
