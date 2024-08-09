package state

import (
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
	logger     logging.Interface

	// Table mapping workspace IDs to states
	cache *resource.Table[*State]

	*pubsub.Broker[*State]
}

type ServiceOptions struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Tasks      *task.Service
	Logger     logging.Interface
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*State](opts.Logger)
	return &Service{
		modules:    opts.Modules,
		workspaces: opts.Workspaces,
		tasks:      opts.Tasks,
		cache:      resource.NewTable(broker),
		Broker:     broker,
		logger:     opts.Logger,
	}
}

// Get retrieves the state for a workspace.
func (s *Service) Get(workspaceID resource.ID) (*State, error) {
	return s.cache.Get(workspaceID)
}

// GetResource retrieves a state resource.
//
// TODO: this is massively inefficient
func (s *Service) GetResource(resourceID resource.ID) (*Resource, error) {
	for _, state := range s.cache.List() {
		for _, res := range state.Resources {
			if res.ID == resourceID {
				return res, nil
			}
		}
	}
	return nil, resource.ErrNotFound
}

// Reload creates a task to repopulate the local cache of the state of the given
// workspace.
func (s *Service) Reload(workspaceID resource.ID) (task.Spec, error) {
	return s.createTaskSpec(workspaceID, task.Spec{
		Command: []string{"state", "pull"},
		JSON:    true,
		AfterExited: func(t *task.Task) {
			state, err := newState(t.Workspace(), t.NewReader(false))
			if err != nil {
				s.logger.Error("reloading state", "error", err, "workspace", t.Workspace())
				return
			}
			// Skip caching state if identical to already cached state.
			//
			// NOTE: re-caching the same state is harmless, but each re-caching
			// generates an event, which reloads the state in the TUI, which
			// makes for unreliable integration tests....instead the tests can
			// wait for a certain serial to appear and be sure no further
			// updates will be made before checking for content.
			if cached, err := s.cache.Get(workspaceID); err == nil {
				if cached.Serial == state.Serial {
					s.logger.Info("skipping caching of reloaded state: identical serial", "state", state)
					return
				}
			}
			// Add/replace state in cache.
			s.cache.Add(workspaceID, state)
			s.logger.Info("reloaded state", "state", state)
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

func (s *Service) Delete(workspaceID resource.ID, addrs ...ResourceAddress) (task.Spec, error) {
	addrStrings := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrings[i] = string(addr)
	}
	return s.createTaskSpec(workspaceID, task.Spec{
		Blocking: true,
		Command:  []string{"state", "rm"},
		Args:     addrStrings,
		AfterError: func(t *task.Task) {
			s.logger.Error("deleting resources", "error", t.Err, "resources", addrs)
		},
		AfterExited: func(t *task.Task) {
			s.CreateReloadTask(workspaceID)
		},
	})
}

func (s *Service) Taint(workspaceID resource.ID, addr ResourceAddress) (task.Spec, error) {
	return s.createTaskSpec(workspaceID, task.Spec{
		Blocking: true,
		Command:  []string{"taint"},
		Args:     []string{string(addr)},
		AfterError: func(t *task.Task) {
			s.logger.Error("tainting resource", "error", t.Err, "resource", addr)
		},
		AfterExited: func(t *task.Task) {
			s.CreateReloadTask(workspaceID)
		},
	})
}

func (s *Service) Untaint(workspaceID resource.ID, addr ResourceAddress) (task.Spec, error) {
	return s.createTaskSpec(workspaceID, task.Spec{
		Blocking: true,
		Command:  []string{"untaint"},
		Args:     []string{string(addr)},
		AfterError: func(t *task.Task) {
			s.logger.Error("untainting resource", "error", t.Err, "resource", addr)
		},
		AfterExited: func(t *task.Task) {
			s.CreateReloadTask(workspaceID)
		},
	})
}

func (s *Service) Move(workspaceID resource.ID, src, dest ResourceAddress) (task.Spec, error) {
	return s.createTaskSpec(workspaceID, task.Spec{
		Blocking: true,
		Command:  []string{"state", "mv"},
		Args:     []string{string(src), string(dest)},
		AfterError: func(t *task.Task) {
			s.logger.Error("moving resource", "error", t.Err, "resources", src)
		},
		AfterExited: func(t *task.Task) {
			s.CreateReloadTask(workspaceID)
		},
	})
}

// TODO: move this logic into task.Create
func (s *Service) createTaskSpec(workspaceID resource.ID, opts task.Spec) (task.Spec, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return task.Spec{}, err
	}
	opts.Parent = ws
	opts.Env = []string{ws.TerraformEnv()}
	opts.Path = ws.ModulePath()

	return opts, nil
}
