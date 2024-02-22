package workspace

import (
	"context"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

// enqueuer determines which tasks should be added to the global queue for
// processing
type enqueuer struct {
	workspaces *Service
	modules    moduleGetter
	tasks      taskLister
}

type moduleGetter interface {
	Get(path string) (*module.Module, error)
}

type taskLister interface {
	List(opts task.ListOptions) []*task.Task
}

func startEnqueuer(
	ctx context.Context,
	workspaces *Service,
	modules *module.Service,
	tasks *task.Service,
) {

	sub, unsub := tasks.Watch(ctx)
	defer unsub()

	e := enqueuer{
		workspaces: workspaces,
		modules:    modules,
		tasks:      tasks,
	}

	for event := range sub {
		for _, t := range e.Enqueue(event) {
			// TODO: log error
			_, _ = tasks.Enqueue(t.ID)
		}
	}
}

// Enqueue uses a recent task event to nominate tasks for enqueuing.
func (e *enqueuer) Enqueue(event resource.Event[*task.Task]) []*task.Task {
	// Only enqueue tasks if module is initialized
	if mod, err := e.modules.Get(event.Payload.Path); err != nil {
		// log error
		return nil
	} else if mod.Status != module.Initialized {
		return nil
	}

	// Retrieve module's active tasks
	active := e.tasks.List(task.ListOptions{
		Path:   &event.Payload.Path,
		Status: []task.Status{task.Queued, task.Running},
	})
	if len(active) > 0 && active[0].Kind == module.InitTask {
		// An active init task blocks pending tasks
		return nil
	}

	// Retrieve module's pending tasks in order of oldest first.
	pending := e.tasks.List(task.ListOptions{
		Path:   &event.Payload.Path,
		Status: []task.Status{task.Pending},
		Oldest: true,
	})
	if len(pending) == 0 {
		// Nothing to enqueue.
		return nil
	}
	// If the oldest pending task is an init task then enqueue only that task.
	if pending[0].Kind == module.InitTask {
		return pending[0:1]
	}

	// Build a set of blocked workspaces
	blocked := make(map[string]struct{})
	for _, t := range active {
		if t.Workspace == nil {
			continue
		}
		switch t.Kind {
		case ApplyTask, DestroyTask, RefreshTask:
			// An active apply task blocks pending tasks
			blocked[*t.Workspace] = struct{}{}
		}
	}
	// Remove blocked tasks from pending tasks
	for i, t := range pending {
		if t.Workspace == nil {
			// Module-specific task
			switch t.Kind {
			case module.InitTask:
				// Only enqueue pending tasks up to init task (not including
				// init task)
				return pending[:i]
			}
		} else {
			// Workspace-specific task
			if _, ok := blocked[*t.Workspace]; ok {
				// Remove blocked task
				pending = append(pending[:i], pending[i+1:]...)
			} else {
				switch t.Kind {
				case ApplyTask, DestroyTask, RefreshTask:
					// Found blocking task in pending queue; no further tasks shall be
					// enqueued
					blocked[*t.Workspace] = struct{}{}
				}
			}
		}
	}
	// Enqueue pending tasks
	return pending
}
