package module

import (
	"context"
	"slices"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/terraform"
	"golang.org/x/exp/maps"
)

type Service struct {
	broker *pubsub.Broker[*Module]
	tasks  *task.Service
	// Pug working directory
	workdir string
	// Modules keyed by path
	store map[string]*Module
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
	if s.store == nil {
		s.store = make(map[string]*Module, len(found))
	}
	for _, mod := range found {
		// Add module if it isn't stored already
		if _, ok := s.store[mod.Path]; !ok {
			s.store[mod.Path] = mod
		}
	}
	// cleanup store, removing those that are not in found
	for path := range s.store {
		if !slices.ContainsFunc(found, func(m *Module) bool {
			return m.Path == path
		}) {
			_ = s.Delete(path)
		}

	}
	return nil
}

// TODO: InitAll()

// Init invokes terraform init on the module.
func (s *Service) Init(path string) (*Module, *task.Task, error) {
	mod, err := s.Get(path)
	if err != nil {
		return nil, nil, err
	}
	// create asynchronous task that runs terraform init
	tsk, err := s.tasks.Create(task.CreateOptions{
		Kind:      InitTask,
		Parent:    mod,
		Args:      []string{"init", "-input=false"},
		Path:      mod.Path,
		Exclusive: terraform.IsPluginCacheUsed(),
	})
	if err != nil {
		return nil, nil, err
	}
	mod.updateStatus(Initializing)
	go func() {
		if err := tsk.Wait(); err != nil {
			mod.updateStatus(Misconfigured)
		} else {
			mod.updateStatus(Initialized)
		}
	}()
	return mod, tsk, nil
}

func (s *Service) List() []*Module {
	return maps.Values(s.store)
}

func (s *Service) CreateTask(spec task.CreateOptions) (*task.Task, error) {
	return nil, nil
}

func (s *Service) Get(path string) (*Module, error) {
	mod, ok := s.store[path]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return mod, nil
}

func (s *Service) Watch(ctx context.Context) (<-chan resource.Event[*Module], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Delete(path string) error {
	mod, ok := s.store[path]
	if !ok {
		return resource.ErrNotFound
	}
	s.broker.Publish(resource.DeletedEvent, mod)
	delete(s.store, path)
	return nil
}
