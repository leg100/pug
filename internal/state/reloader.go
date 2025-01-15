package state

import (
	"fmt"

	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type reloader struct {
	*Service
}

// Reload creates a task to repopulate the local cache of the state of the given
// workspace.
func (r *reloader) Reload(workspaceID resource.ID) (task.Spec, error) {
	return r.createTaskSpec(workspaceID, task.Spec{
		Execution: task.Execution{
			TerraformCommand: []string{"state", "pull"},
		},
		JSON: true,
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			state, err := newState(workspaceID, t.NewReader(false))
			if err != nil {
				return nil, fmt.Errorf("constructing pug state: %w", err)
			}
			// Skip caching state if identical to already old state.
			//
			// NOTE: re-caching the same state is harmless, but each re-caching
			// generates an event, which reloads the state in the TUI, which
			// makes for flaky integration tests....instead the tests can
			// wait for a certain serial to appear and be sure no further
			// updates will be made before checking for content.
			old, err := r.cache.Get(workspaceID)
			if err == nil && old.Serial == state.Serial {
				return newReloadSummary(old, state), nil
			}
			// Add/replace state in cache.
			r.cache.Add(workspaceID, state)
			return newReloadSummary(old, state), nil
		},
	})
}

func (s *Service) CreateReloadTask(workspaceID resource.ID) (*task.Task, error) {
	spec, err := s.Reload(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("creating reload task spec: %w", err)
	}
	task, err := s.tasks.Create(spec)
	if err != nil {
		return nil, fmt.Errorf("creating reload task: %w", err)
	}
	return task, nil
}
