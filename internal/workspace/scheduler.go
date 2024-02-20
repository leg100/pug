package workspace

import (
	"sync"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

// scheduler schedules tasks
type scheduler struct {
	// module tasks keyed by module path
	cache map[ID]*task.Categories
	mu    sync.Mutex

	modscheduler *module.Scheduler
	service      *Service
	tasks        *task.Service
}

// Handle updates the task cache and then makes a decision as to whether tasks
// should be enqueued.
func (s *scheduler) Handle(event resource.Event[*task.Task]) []*task.Task {
	wsname := event.Payload.WorkspaceName
	if wsname == nil {
		// TODO: log error
		return nil
	}
	id := ID{
		Path: event.Payload.Path,
		Name: *wsname,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update cache
	categories := s.cache[id]
	categories.Categorize(event)
	s.cache[id] = categories

	// Pass event to module scheduler too and let it decide first which tasks
	// should be enqueued: both the module and workspace schedulers must agree
	// on which tasks should be enqueued.
	enqueue := s.modscheduler.Handle(event)
	if len(enqueue) == 0 {
		// Module scheduler decision that no tasks should be scheduled takes
		// precedence
		return nil
	}
	// Now workspace scheduler must agree on the tasks the module scheduler has
	// nominated for enqueuing.
	if len(categories.Active()) > 0 {
		switch categories.Active()[0].Kind {
		case ApplyTask, DestroyTask, RefreshTask:
			// An active ApplyTask blocks pending task
			return nil
		}
	}
	// Enqueue all tasks the module schedule nominated for enqueuing.
	return enqueue
}
