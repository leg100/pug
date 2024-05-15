package module

import (
	"fmt"
	"slices"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	table       *resource.Table[*Module]
	tasks       *task.Service
	workdir     internal.Workdir
	pluginCache bool
	logger      logging.Interface

	*pubsub.Broker[*Module]
}

type ServiceOptions struct {
	TaskService *task.Service
	Workdir     internal.Workdir
	PluginCache bool
	Logger      logging.Interface
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Module](opts.Logger)
	table := resource.NewTable(broker)

	opts.Logger.AddEnricher(&logEnricher{table: table})

	return &Service{
		table:       table,
		Broker:      broker,
		tasks:       opts.TaskService,
		workdir:     opts.Workdir,
		pluginCache: opts.PluginCache,
		logger:      opts.Logger,
	}
}

// Reload searches the working directory recursively for modules and adds them
// to the store before pruning those that are currently stored but can no longer
// be found.
func (s *Service) Reload() (added []string, removed []string, err error) {
	found, err := findModules(s.logger, s.workdir)
	if err != nil {
		return nil, nil, err
	}
	for _, path := range found {
		// Add module if it isn't in pug already
		if _, err := s.GetByPath(path); err == resource.ErrNotFound {
			mod := New(s.workdir, path)
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
	s.logger.Info("reloaded modules", "added", added, "removed", removed)
	return
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
			s.table.Update(moduleID, func(mod *Module) error {
				mod.InitInProgress = true
				return nil
			})
		},
		AfterExited: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) error {
				mod.Initialized = internal.Bool(true)
				mod.InitInProgress = false
				return nil
			})
		},
		AfterError: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) error {
				mod.Initialized = internal.Bool(false)
				mod.InitInProgress = false
				return nil
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

func (s *Service) SetCurrent(moduleID, workspaceID resource.ID) error {
	_, err := s.table.Update(moduleID, func(existing *Module) error {
		existing.CurrentWorkspaceID = &workspaceID
		return nil
	})
	if err != nil {
		s.logger.Error("setting current workspace", "module", moduleID, "workspace", workspaceID, "error", err)
		return err
	}
	s.logger.Debug("set current workspace", "module", moduleID, "workspace", workspaceID)
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
			s.table.Update(moduleID, func(mod *Module) error {
				mod.FormatInProgress = true
				return nil
			})
		},
		AfterExited: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) error {
				mod.Formatted = internal.Bool(true)
				mod.FormatInProgress = false
				return nil
			})
		},
		AfterError: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) error {
				mod.Formatted = internal.Bool(false)
				mod.FormatInProgress = false
				return nil
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
			s.table.Update(moduleID, func(mod *Module) error {
				mod.ValidationInProgress = true
				return nil
			})
		},
		AfterExited: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) error {
				mod.Valid = internal.Bool(true)
				mod.ValidationInProgress = false
				return nil
			})
		},
		AfterError: func(*task.Task) {
			s.table.Update(moduleID, func(mod *Module) error {
				mod.Valid = internal.Bool(false)
				mod.ValidationInProgress = false
				return nil
			})
		},
	})
}

// TODO: move this logic into task.Create
func (s *Service) CreateTask(mod *Module, opts task.CreateOptions) (*task.Task, error) {
	opts.Parent = mod
	opts.Path = mod.Path
	return s.tasks.Create(opts)
}
