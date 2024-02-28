package module

import (
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/terraform"
	"golang.org/x/exp/maps"
)

type Service struct {
	broker *pubsub.Broker[*Module]
	tasks  *task.Service

	workdir string

	modules map[uuid.UUID]*Module
	mu      sync.Mutex
}

type ServiceOptions struct {
	TaskService *task.Service
	Workdir     string
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Module]()
	return &Service{
		tasks:   opts.TaskService,
		workdir: opts.Workdir,
		broker:  broker,
	}
}

// Reload searches the working directory recursively for modules and adds them
// to the store before pruning those that are currently stored but can no longer
// be found.
func (s *Service) Reload() error {
	found, err := findModules(s.workdir)
	if err != nil {
		return err
	}
	if s.modules == nil {
		s.modules = make(map[uuid.UUID]*Module, len(found))
	}
	for _, path := range found {
		// Add module if it isn't in pug already
		if _, err := s.GetByPath(path); err == resource.ErrNotFound {
			mod := &Module{Resource: resource.New(nil), Path: path}
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
func (s *Service) Init(id uuid.UUID) (*Module, *task.Task, error) {
	mod, err := s.Get(id)
	if err != nil {
		return nil, nil, err
	}
	// create asynchronous task that runs terraform init
	tsk, err := s.tasks.Create(task.CreateOptions{
		Parent:    mod.Resource,
		Path:      mod.Path,
		Command:   []string{"init"},
		Args:      []string{"-input=false"},
		Blocking:  true,
		Exclusive: terraform.IsPluginCacheUsed(),
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

func (s *Service) Get(id uuid.UUID) (*Module, error) {
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

func (s *Service) Watch(ctx context.Context) (<-chan resource.Event[*Module], func()) {
	return s.broker.Subscribe(ctx)
}
