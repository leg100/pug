package state

import (
	"fmt"
	"log/slog"

	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type reloader struct {
	*Service
}

type ReloadSummary struct {
	Old, New *State
}

func (s ReloadSummary) OldSerial() int {
	oldSerial := -1
	if s.Old != nil {
		oldSerial = int(s.Old.Serial)
	}
	return oldSerial
}

func (s ReloadSummary) NewSerial() int {
	newSerial := -1
	if s.New != nil {
		newSerial = int(s.New.Serial)
	}
	return newSerial
}

func (s ReloadSummary) String() string {
	return fmt.Sprintf("#%d->#%d", s.OldSerial(), s.NewSerial())
}

func (s ReloadSummary) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("old.state.serial", s.OldSerial()),
		slog.Any("new.state.serial", s.NewSerial()),
	)
}

// Reload creates a task to repopulate the local cache of the state of the given
// workspace.
func (r *reloader) Reload(workspaceID resource.ID) (task.Spec, error) {
	return r.createTaskSpec(workspaceID, task.Spec{
		Command: []string{"state", "pull"},
		JSON:    true,
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			state, err := newState(t.Workspace(), t.NewReader(false))
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
				return ReloadSummary{Old: old, New: state}, nil
			}
			// Add/replace state in cache.
			r.cache.Add(workspaceID, state)
			return ReloadSummary{Old: old, New: state}, nil
		},
	})
}

func (s *Service) CreateReloadTask(workspaceID resource.ID) (*task.Task, error) {
	spec, err := s.Reload(workspaceID)
	if err != nil {
		return nil, err
	}
	return s.tasks.Create(spec)
}
