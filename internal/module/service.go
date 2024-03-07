package module

import (
	"context"
	"slices"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	table  *resource.Table[*Module]
	broker *pubsub.Broker[*Module]

	tasks *task.Service

	workdir     string
	pluginCache bool
}

type ServiceOptions struct {
	TaskService *task.Service
	Workdir     string
	PluginCache bool
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Module]()
	svc := &Service{
		table:       resource.NewTable[*Module](broker),
		broker:      broker,
		tasks:       opts.TaskService,
		workdir:     opts.Workdir,
		pluginCache: opts.PluginCache,
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
			mod := New(path)
			s.table.Add(mod.ID, mod)
		}
	}

	// Cleanup existing modules, removing those that are no longer to be found
	for _, existing := range s.table.List() {
		if !slices.Contains(found, existing.Path()) {
			s.table.Delete(existing.ID)
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
		Path:     mod.Path(),
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
	return s.table.List()
}

func (s *Service) Get(id resource.ID) (*Module, error) {
	return s.table.Get(id)
}

func (s *Service) GetByPath(path string) (*Module, error) {
	for _, mod := range s.table.List() {
		if path == mod.Path() {
			return mod, nil
		}
	}
	return nil, resource.ErrNotFound
}

func (s *Service) Subscribe(ctx context.Context) (<-chan resource.Event[*Module], func()) {
	return s.broker.Subscribe(ctx)
}
