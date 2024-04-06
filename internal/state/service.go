package state

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	modules    *module.Service
	workspaces *workspace.Service
	tasks      *task.Service
	broker     *pubsub.Broker[*State]
	logger     *logging.Logger

	// Table mapping workspace IDs to states
	cache *resource.Table[*State]
}

type ServiceOptions struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	TaskService      *task.Service
	Logger           *logging.Logger
}

func NewService(ctx context.Context, opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*State]()
	svc := &Service{
		modules:    opts.ModuleService,
		workspaces: opts.WorkspaceService,
		tasks:      opts.TaskService,
		cache:      resource.NewTable(broker),
		broker:     broker,
		logger:     opts.Logger,
	}

	// Whenever a workspace is added, pull its state
	go func() {
		for event := range opts.WorkspaceService.Subscribe(ctx) {
			if event.Type == resource.CreatedEvent {
				_, _ = svc.Reload(event.Payload.ID)
			}
		}
	}()

	return svc
}

// Get retrieves the state for a workspace.
func (s *Service) Get(workspaceID resource.ID) (*State, error) {
	state, err := s.cache.Get(workspaceID)
	if errors.Is(err, resource.ErrNotFound) {
		return EmptyState(workspaceID), nil
	}
	return state, err
}

// Reload creates a task to repopulate the local cache of the state of the given
// workspace.
func (s *Service) Reload(workspaceID resource.ID) (*task.Task, error) {
	err := s.updateStateStatus(workspaceID, func(existing *State) error {
		return existing.startReload()
	})
	if err != nil {
		return nil, err
	}

	revertIdle := func() {
		s.updateStateStatus(workspaceID, func(existing *State) error {
			existing.State = IdleState
			return nil
		})
	}

	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, err
	}

	task, err := s.createTask(workspaceID, task.CreateOptions{
		Command: []string{"show", "-json"},
		JSON:    true,
		AfterError: func(t *task.Task) {
			s.logger.Error("reloading state", "error", t.Err, "workspace", ws)
		},
		AfterExited: func(t *task.Task) {
			var file StateFile
			if err := json.NewDecoder(t.NewReader()).Decode(&file); err != nil {
				s.logger.Error("reloading state", "error", err, "workspace", ws)
				return
			}
			current := NewState(workspaceID, file)
			// For each current resource, check if it previously existed in the
			// cache, and if so, copy across its status.
			s.cache.Update(workspaceID, func(previous *State) error {
				for currentAddress := range current.Resources {
					if previousResource, ok := previous.Resources[currentAddress]; ok {
						current.Resources[currentAddress].Status = previousResource.Status
					}
				}
				return nil
			})
			// table.Add replaces state if it exists already, which is what we
			// want.
			s.cache.Add(workspaceID, current)
			s.logger.Info("reloaded state", "workspace", ws, "resources", len(current.Resources))
		},
		AfterFinish: func(t *task.Task) {
			revertIdle()
		},
	})
	if err != nil {
		revertIdle()
		return nil, err
	}
	return task, nil
}

func (s *Service) Delete(workspaceID resource.ID, addrs ...ResourceAddress) (*task.Task, error) {
	addrStrings := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrings[i] = string(addr)
	}
	return s.createTask(workspaceID, task.CreateOptions{
		Blocking: true,
		Command:  []string{"state", "rm"},
		Args:     addrStrings,
		AfterCreate: func(t *task.Task) {
			s.updateResourceStatus(workspaceID, Removing, addrs...)
		},
		AfterError: func(t *task.Task) {
			s.updateResourceStatus(workspaceID, Idle, addrs...)
			s.logger.Error("deleting resources", "error", t.Err, "resources", addrs)
		},
		AfterCanceled: func(t *task.Task) {
			s.updateResourceStatus(workspaceID, Idle, addrs...)
		},
		AfterExited: func(t *task.Task) {
			s.cache.Update(workspaceID, func(existing *State) error {
				// Remove resources from cache
				for _, addr := range addrs {
					delete(existing.Resources, addr)
				}
				return nil
			})
		},
	})
}

func (s *Service) Taint(workspaceID resource.ID, addr ResourceAddress) (*task.Task, error) {
	return s.createTask(workspaceID, task.CreateOptions{
		Blocking: true,
		Command:  []string{"taint"},
		Args:     []string{string(addr)},
		AfterCreate: func(t *task.Task) {
			s.updateResourceStatus(workspaceID, Tainting, addr)
		},
		AfterError: func(t *task.Task) {
			s.updateResourceStatus(workspaceID, Idle, addr)
			s.logger.Error("tainting resources", "error", t.Err, "resources", addr)
		},
		AfterCanceled: func(t *task.Task) {
			s.updateResourceStatus(workspaceID, Idle, addr)
		},
		AfterExited: func(t *task.Task) {
			s.updateResourceStatus(workspaceID, Tainted, addr)
		},
	})
}

func (s *Service) Subscribe(ctx context.Context) <-chan resource.Event[*State] {
	return s.broker.Subscribe(ctx)
}

func (s *Service) createTask(workspaceID resource.ID, opts task.CreateOptions) (*task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, err
	}
	opts.Parent = ws.Resource
	opts.Env = []string{ws.TerraformEnv()}

	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return nil, err
	}
	opts.Path = mod.Path

	return s.tasks.Create(opts)
}

func (s *Service) updateStateStatus(workspaceID resource.ID, fn func(*State) error) error {
	var err error
	s.cache.Update(workspaceID, func(existing *State) error {
		if updateErr := fn(existing); updateErr != nil {
			err = updateErr
		}
		return nil
	})
	return err
}

func (s *Service) updateResourceStatus(workspaceID resource.ID, state ResourceStatus, addrs ...ResourceAddress) {
	s.cache.Update(workspaceID, func(existing *State) error {
		for _, res := range existing.Resources {
			if slices.Contains(addrs, res.Address) {
				res.Status = state
			}
		}
		return nil
	})
}
