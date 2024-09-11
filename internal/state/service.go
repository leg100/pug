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
	*reloader
}

type ServiceOptions struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Tasks      *task.Service
	Logger     logging.Interface
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*State](opts.Logger)
	s := &Service{
		modules:    opts.Modules,
		workspaces: opts.Workspaces,
		tasks:      opts.Tasks,
		cache:      resource.NewTable(broker),
		Broker:     broker,
		logger:     opts.Logger,
	}
	s.reloader = &reloader{s}
	return s
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

func (s *Service) Delete(workspaceID resource.ID, addrs ...ResourceAddress) (task.Spec, error) {
	addrStrings := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrings[i] = string(addr)
	}
	return s.createTaskSpec(workspaceID, task.Spec{
		Blocking: true,
		Execution: task.Execution{
			TerraformCommand: []string{"state", "rm"},
			Args:             addrStrings,
		},
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
		Execution: task.Execution{
			TerraformCommand: []string{"taint"},
			Args:             []string{string(addr)},
		},
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
		Execution: task.Execution{
			TerraformCommand: []string{"untaint"},
			Args:             []string{string(addr)},
		},
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
		Execution: task.Execution{
			TerraformCommand: []string{"state", "mv"},
			Args:             []string{string(src), string(dest)},
		},
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
	mod, err := s.modules.Get(ws.ModuleID)
	if err != nil {
		return task.Spec{}, err
	}
	opts.ModuleID = &mod.ID
	opts.WorkspaceID = &ws.ID
	opts.Env = []string{ws.TerraformEnv()}
	opts.Path = mod.Path

	return opts, nil
}
