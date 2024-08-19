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

// LoadStateUponWorkspaceLoad automatically loads state for a workspace when it
// is loaded into pug for the first time.
func (s *Service) LoadStateUponWorkspaceLoad(sub <-chan resource.Event[*workspace.Workspace]) {
	for event := range sub {
		if event.Type == resource.CreatedEvent {
			_, _ = s.CreateReloadTask(event.Payload.ID)
		}
	}
}

// LoadStateUponInit automatically loads state whenever a module is successfully
// initialised, and the state for a workspace belonging to the module has not
// been loaded yet.
func (s *Service) LoadStateUponInit(sub <-chan resource.Event[*task.Task]) {
	for event := range sub {
		if !module.IsInitTask(event.Payload) {
			continue
		}
		if event.Payload.State != task.Exited {
			continue
		}
		opts := workspace.ListOptions{ModuleID: event.Payload.Module().GetID()}
		workspaces := s.workspaces.List(opts)
		for _, ws := range workspaces {
			if _, err := s.cache.Get(ws.ID); err == nil {
				// State already loaded
				continue
			}
			_, _ = s.CreateReloadTask(ws.ID)
		}
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
