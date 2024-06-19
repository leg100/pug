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
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	TaskService      *task.Service
	Logger           logging.Interface
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*State](opts.Logger)
	svc := &Service{
		modules:    opts.ModuleService,
		workspaces: opts.WorkspaceService,
		tasks:      opts.TaskService,
		cache:      resource.NewTable(broker),
		Broker:     broker,
		logger:     opts.Logger,
	}

	// Whenever a workspace is added, pull its state
	go func() {
		for event := range opts.WorkspaceService.Subscribe() {
			if event.Type == resource.CreatedEvent {
				_, _ = svc.Reload(event.Payload.ID)
			}
		}
	}()

	return svc
}

// Get retrieves the state for a workspace.
func (s *Service) Get(workspaceID resource.ID) (*State, error) {
	return s.cache.Get(workspaceID)
}

// Reload creates a task to repopulate the local cache of the state of the given
// workspace.
func (s *Service) Reload(workspaceID resource.ID) (*task.Task, error) {
	return s.createTask(workspaceID, task.CreateOptions{
		Command: []string{"state", "pull"},
		JSON:    true,
		AfterExited: func(t *task.Task) {
			state, err := newState(t.Workspace(), t.NewReader())
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

func (s *Service) Delete(workspaceID resource.ID, addrs ...ResourceAddress) (*task.Task, error) {
	addrStrings := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrings[i] = string(addr)
	}
	return s.createTask(workspaceID, task.CreateOptions{
		Blocking: true,
		Command:  []string{"state", "rm"},
		Args:     addrStrings,
		AfterError: func(t *task.Task) {
			s.logger.Error("deleting resources", "error", t.Err, "resources", addrs)
		},
		AfterExited: func(t *task.Task) {
			s.Reload(workspaceID)
		},
	})
}

func (s *Service) Taint(workspaceID resource.ID, addr ResourceAddress) (*task.Task, error) {
	return s.createTask(workspaceID, task.CreateOptions{
		Blocking: true,
		Command:  []string{"taint"},
		Args:     []string{string(addr)},
		AfterError: func(t *task.Task) {
			s.logger.Error("tainting resource", "error", t.Err, "resource", addr)
		},
		AfterExited: func(t *task.Task) {
			s.Reload(workspaceID)
		},
	})
}

func (s *Service) Untaint(workspaceID resource.ID, addr ResourceAddress) (*task.Task, error) {
	return s.createTask(workspaceID, task.CreateOptions{
		Blocking: true,
		Command:  []string{"untaint"},
		Args:     []string{string(addr)},
		AfterError: func(t *task.Task) {
			s.logger.Error("untainting resource", "error", t.Err, "resource", addr)
		},
		AfterExited: func(t *task.Task) {
			s.Reload(workspaceID)
		},
	})
}

func (s *Service) Move(workspaceID resource.ID, src, dest ResourceAddress) (*task.Task, error) {
	return s.createTask(workspaceID, task.CreateOptions{
		Blocking: true,
		Command:  []string{"state", "mv"},
		Args:     []string{string(src), string(dest)},
		AfterError: func(t *task.Task) {
			s.logger.Error("moving resource", "error", t.Err, "resources", src)
		},
		AfterExited: func(t *task.Task) {
			s.Reload(workspaceID)
		},
	})
}

// TODO: move this logic into task.Create
func (s *Service) createTask(workspaceID resource.ID, opts task.CreateOptions) (*task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, err
	}
	opts.Parent = ws
	opts.Env = []string{ws.TerraformEnv()}
	opts.Path = ws.ModulePath()

	return s.tasks.Create(opts)
}
