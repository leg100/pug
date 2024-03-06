package module

import (
	"context"
	"slices"
	"sync"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"golang.org/x/exp/maps"
)

type Service struct {
	broker *pubsub.Broker[*Module]
	tasks  *task.Service

	workdir     string
	pluginCache bool

	modules map[resource.ID]*Module
	mu      sync.Mutex
}

type ServiceOptions struct {
	TaskService *task.Service
	Workdir     string
	PluginCache bool
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Module]()
	svc := &Service{
		tasks:       opts.TaskService,
		workdir:     opts.Workdir,
		broker:      broker,
		pluginCache: opts.PluginCache,
		modules:     make(map[resource.ID]*Module),
	}
	return svc
}

// Reload searches the working directory recursively for modules and adds them
// to the store before pruning those that are currently stored but can no longer
// be found.
func (s *Service) Reload() error {
	found, err := findModules(s.workdir)
	if err != nil {
		return err
	}
	for _, path := range found {
		// Add module if it isn't in pug already
		if _, err := s.GetByPath(path); err == resource.ErrNotFound {
			mod := &Module{
				Resource: resource.New(resource.Module, path, nil),
				Path:     path,
			}
			s.mu.Lock()
			s.modules[mod.ID] = mod
			s.mu.Unlock()
			s.broker.Publish(resource.CreatedEvent, mod)
		}
	}

	// Cleanup existing modules, removing those that are no longer to be found
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, existing := range s.modules {
		if !slices.Contains(found, existing.Path) {
			s.broker.Publish(resource.DeletedEvent, existing)
			delete(s.modules, id)
		}
	}
	return nil
}

// Init invokes terraform init on the module.
func (s *Service) Init(moduleID resource.ID) (*Module, *task.Task, error) {
	mod, err := s.Get(moduleID)
	if err != nil {
		return nil, nil, err
	}
	// create asynchronous task that runs terraform init
	tsk, err := s.tasks.Create(task.CreateOptions{
		Parent:   mod.Resource,
		Path:     mod.Path,
		Command:  []string{"init"},
		Args:     []string{"-input=false"},
		Blocking: true,
		// The terraform plugin cache is not concurrency-safe, so only allow one
		// init task to run at any given time.
		Exclusive: s.pluginCache,
		AfterCreate: func(*task.Task) {
			// log
		},
		AfterExited: func(*task.Task) {
			// log
		},
		AfterError: func(*task.Task) {
			// log error
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return mod, tsk, nil
}

func (s *Service) List() []*Module {
	s.mu.Lock()
	defer s.mu.Unlock()

	return maps.Values(s.modules)
}

func (s *Service) Get(id resource.ID) (*Module, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mod, ok := s.modules[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return mod, nil
}

func (s *Service) GetByPath(path string) (*Module, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, mod := range s.modules {
		if path == mod.Path {
			return mod, nil
		}
	}
	return nil, resource.ErrNotFound
}

func (s *Service) Subscribe(ctx context.Context) (<-chan resource.Event[*Module], func()) {
	return s.broker.Subscribe(ctx)
}
