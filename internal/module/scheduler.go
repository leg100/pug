package module

import (
	"sync"

	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

// Scheduler schedules tasks
type Scheduler struct {
	// module tasks keyed by module path
	cache map[string]*task.Categories
	mu    sync.Mutex

	service *Service
	tasks   *task.Service
}

func (s *Scheduler) Handle(event resource.Event[*task.Task]) []*task.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	categories := s.cache[event.Payload.Path]
	categories.Categorize(event)
	s.cache[event.Payload.Path] = categories

	mod, err := s.service.Get(event.Payload.Path)
	if err != nil {
		// log error
		return nil
	}

	if mod.Status != Initialized {
		// Cannot enqueue tasks for a module that is not in the initialized
		// state.
		return nil
	}
	if len(categories.Active()) > 0 && categories.Active()[0].Kind == InitTask {
		// A queued InitTask blocks pending task (and a queue for a module can
		// only contain *one* InitTask, so we're safe just checking the first
		// item).
		return nil
	}
	var enqueue []*task.Task
	for _, pending := range categories.Pending {
		if pending.Kind != InitTask {
			enqueue = append(enqueue, pending)
		}
	}
	return enqueue
}
