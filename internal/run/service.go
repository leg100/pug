package run

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	tasks      *task.Service
	workspaces *workspace.Service
	modules    *module.Service
	// Runs keyed by run ID
	runs map[uuid.UUID]*Run
	// Mutex for concurrent read/write of runs
	mu     sync.Mutex
	broker *pubsub.Broker[*Run]
}

// Create a run, triggering a plan task.
func (s *Service) Create(id workspace.ID, opts CreateOptions) (*Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, err := s.workspaces.Get(id)
	if err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}
	mod, err := s.modules.Get(id.Path)
	if err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}
	run, err := newRun(ws, mod, opts)
	if err != nil {
		return nil, fmt.Errorf("creating run: %w", err)
	}
	s.broker.Publish(resource.CreatedEvent, run)
	return run, nil
}

// Apply triggers an apply for a run. The run must be in the planned state.
func (s *Service) Apply(id uuid.UUID) (*Run, *task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id(path, name)]
	if !ok {
		return nil, nil, resource.ErrNotFound
	}
	return ws, nil, nil
}
