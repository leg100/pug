package module

import (
	"context"
	"fmt"
	"slices"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/terraform"
	"golang.org/x/exp/maps"
)

type Service struct {
	broker  *pubsub.Broker[*Module]
	tasks   *task.Service
	modules map[resource.Resource]*Module
	workdir string
	*factory
	// TODO: Mutex for making atomic changes to modules and/or manipulating the
	// modules map concurrently
}

type ServiceOptions struct {
	TaskService *task.Service
	Workdir     string
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Module]()
	factory := &factory{
		// Whenever state is updated, publish it as an event.
		callback: func(t *Module) {
			broker.Publish(resource.UpdatedEvent, t)
		},
	}
	return &Service{
		tasks:   opts.TaskService,
		workdir: opts.Workdir,
		factory: factory,
		broker:  broker,
	}
}

// Reload searches the working directory recursively for modules and adds them
// to the store before pruning those that are currently stored but can no longer
// be found.
func (s *Service) Reload() error {
	found, err := s.findModules(s.workdir)
	if err != nil {
		return err
	}
	if s.modules == nil {
		s.modules = make(map[resource.Resource]*Module, len(found))
	}
	for _, mod := range found {
		// Add module if it isn't stored already
		if _, err := s.GetByPath(mod.Path); err == resource.ErrNotFound {
			s.modules[mod.Resource] = mod
		}
	}
	// Cleanup existing modules, removing those that are not in found
	for _, existing := range s.modules {
		if !slices.ContainsFunc(found, func(m *Module) bool {
			return m.Path == existing.Path
		}) {
			_ = s.Delete(existing.Resource)
		}
	}
	return nil
}

// Init invokes terraform init on the module.
func (s *Service) Init(id resource.Resource) (*Module, *task.Task, error) {
	mod, err := s.Get(id)
	if err != nil {
		return nil, nil, err
	}
	// create asynchronous task that runs terraform init
	tsk, err := s.CreateTask(id, task.CreateOptions{
		Command: []string{"init"},
		Args:    []string{"-input=false"},
		// init blocks module and workspace tasks
		Blocking:  true,
		Exclusive: terraform.IsPluginCacheUsed(),
		AfterCreate: func(*task.Task) {
			mod.updateStatus(Initializing)
		},
		AfterExited: func(*task.Task) {
			mod.updateStatus(Initialized)
		},
		AfterError: func(*task.Task) {
			mod.updateStatus(Misconfigured)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return mod, tsk, nil
}

func (s *Service) List() []*Module {
	return maps.Values(s.modules)
}

const MetadataKey = "module"

// Create module task.
func (s *Service) CreateTask(id resource.Resource, opts task.CreateOptions) (*task.Task, error) {
	// ensure module exists
	mod, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("retrieving module: %s: %w", id, err)
	}
	if opts.Metadata == nil {
		opts.Metadata = make(map[string]string)
	}
	opts.Metadata[MetadataKey] = id
	opts.Path = mod.Path
	return s.tasks.Create(opts)
}

func (s *Service) Get(id resource.Resource) (*Module, error) {
	mod, ok := s.modules[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return mod, nil
}

func (s *Service) GetByPath(path string) (*Module, error) {
	for _, mod := range s.modules {
		if path == mod.Path {
			return mod, nil
		}
	}
	return nil, resource.ErrNotFound
}

func (s *Service) Watch(ctx context.Context) (<-chan resource.Event[*Module], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Delete(id resource.Resource) error {
	mod, ok := s.modules[id]
	if !ok {
		return resource.ErrNotFound
	}
	s.broker.Publish(resource.DeletedEvent, mod)
	delete(s.modules, id)
	return nil
}
