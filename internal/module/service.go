package module

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/leg100/pug/internal"
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
		table:       resource.NewTable(broker),
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
	slog.Info("reloading modules")

	var (
		added   []string
		removed []string
	)

	found, err := findModules(s.workdir)
	if err != nil {
		return err
	}
	for _, path := range found {
		// Add module if it isn't in pug already
		if _, err := s.GetByPath(path); err == resource.ErrNotFound {
			mod := New(path)
			s.table.Add(mod.ID, mod)
			added = append(added, path)
		}
	}

	// Cleanup existing modules, removing those that are no longer to be found
	for _, existing := range s.table.List() {
		if !slices.Contains(found, existing.Path) {
			s.table.Delete(existing.ID)
			removed = append(removed, existing.Path)
		}
	}
	slog.Info("reloaded modules", "added", added, "removed", removed)
	return nil
}

// Init invokes terraform init on the module.
func (s *Service) Init(moduleID resource.ID) (*task.Task, error) {
	mod, err := s.Get(moduleID)
	if err != nil {
		return nil, fmt.Errorf("initializing module: %w", err)
	}
	// create asynchronous task that runs terraform init
	tsk, err := s.CreateTask(mod, task.CreateOptions{
		Command:  []string{"init"},
		Args:     []string{"-input=false"},
		Blocking: true,
		// The terraform plugin cache is not concurrency-safe, so only allow one
		// init task to run at any given time.
		Exclusive: s.pluginCache,
		AfterCreate: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.InitInProgress = true
			})
		},
		AfterExited: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.Initialized = internal.Bool(true)
				mod.InitInProgress = false
			})
		},
		AfterError: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.Initialized = internal.Bool(false)
				mod.InitInProgress = false
			})
		},
	})
	if err != nil {
		return nil, err
	}
	return tsk, nil
}

func (s *Service) List() []*Module {
	return s.table.List()
}

func (s *Service) Get(id resource.ID) (*Module, error) {
	return s.table.Get(id)
}

func (s *Service) GetByPath(path string) (*Module, error) {
	for _, mod := range s.table.List() {
		if path == mod.Path {
			return mod, nil
		}
	}
	return nil, resource.ErrNotFound
}

func (s *Service) Subscribe(ctx context.Context) <-chan resource.Event[*Module] {
	return s.broker.Subscribe(ctx)
}

func (s *Service) SetCurrent(moduleID resource.ID, workspace resource.ID) error {
	err := s.table.Update(moduleID, func(existing *Module) {
		existing.CurrentWorkspaceID = &workspace
	})
	if err != nil {
		return fmt.Errorf("setting current workspace for module: %w", err)
	}
	// TODO: create task to invoke `terraform workspace select <workspace_name>`
	return nil
}

func (s *Service) Format(moduleID resource.ID) (*task.Task, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return nil, fmt.Errorf("formatting module: %w", err)
	}

	return s.CreateTask(mod, task.CreateOptions{
		Command: []string{"fmt"},
		AfterCreate: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.FormatInProgress = true
			})
		},
		AfterExited: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.Formatted = internal.Bool(true)
				mod.FormatInProgress = false
			})
		},
		AfterError: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.Formatted = internal.Bool(false)
				mod.FormatInProgress = false
			})
		},
	})
}

func (s *Service) Validate(moduleID resource.ID) (*task.Task, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return nil, fmt.Errorf("validating module: %w", err)
	}

	return s.CreateTask(mod, task.CreateOptions{
		Command: []string{"validate"},
		AfterCreate: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.ValidationInProgress = true
			})
		},
		AfterExited: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.Valid = internal.Bool(true)
				mod.ValidationInProgress = false
			})
		},
		AfterError: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) {
				mod.Valid = internal.Bool(false)
				mod.ValidationInProgress = false
			})
		},
	})
}

func (s *Service) CreateTask(mod *Module, opts task.CreateOptions) (*task.Task, error) {
	opts.Parent = mod.Resource
	opts.Path = mod.Path
	return s.tasks.Create(opts)
}
