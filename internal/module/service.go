package module

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	table       moduleTable
	tasks       taskCreator
	workdir     internal.Workdir
	pluginCache bool
	logger      logging.Interface
	terragrunt  bool

	*pubsub.Broker[*Module]
}

type ServiceOptions struct {
	Tasks       *task.Service
	Workdir     internal.Workdir
	PluginCache bool
	Logger      logging.Interface
	Terragrunt  bool
}

type taskCreator interface {
	Create(spec task.CreateOptions) (*task.Task, error)
}

type moduleTable interface {
	Add(id resource.ID, row *Module)
	Update(id resource.ID, updater func(existing *Module) error) (*Module, error)
	Delete(id resource.ID)
	Get(id resource.ID) (*Module, error)
	List() []*Module
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Module](opts.Logger)
	table := resource.NewTable(broker)

	opts.Logger.AddEnricher(&logEnricher{table: table})

	return &Service{
		table:       table,
		Broker:      broker,
		tasks:       opts.Tasks,
		workdir:     opts.Workdir,
		pluginCache: opts.PluginCache,
		logger:      opts.Logger,
		terragrunt:  opts.Terragrunt,
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
	for _, opts := range found {
		// Add module if it isn't in pug already, otherwise update in-place
		if mod, err := s.GetByPath(opts.Path); errors.Is(err, resource.ErrNotFound) {
			mod := New(s.workdir, opts)
			s.table.Add(mod.ID, mod)
			added = append(added, opts.Path)
		} else if err != nil {
			return nil, nil, fmt.Errorf("retrieving module: %w", err)
		} else {
			// Update in-place; the backend may have changed.
			s.table.Update(mod.ID, func(existing *Module) error {
				existing.Backend = opts.Backend
				return nil
			})
		}
	}

	// Cleanup existing modules, removing those that are no longer to be found
	for _, existing := range s.table.List() {
		var keep bool
		for _, opts := range found {
			if opts.Path == existing.Path {
				keep = true
				break
			}
		}
		if !keep {
			s.table.Delete(existing.ID)
			removed = append(removed, existing.Path)
		}
	}
	s.logger.Info("reloaded modules", "added", added, "removed", removed)

	if s.terragrunt {
		if err := s.loadTerragruntDependencies(); err != nil {
			s.logger.Error("loading terragrunt dependencies: %w", err)
		}
	}
	return
}

func (s *Service) loadTerragruntDependencies() error {
	task, err := s.tasks.Create(task.CreateOptions{
		Parent:  resource.GlobalResource,
		Command: []string{"graph-dependencies"},
		Wait:    true,
	})
	if err != nil {
		return err
	}
	return s.loadTerragruntDependenciesFromDigraph(task.NewReader(false))
}

func (s *Service) loadTerragruntDependenciesFromDigraph(r io.Reader) error {
	results, err := parseTerragruntGraph(r)
	if err != nil {
		return fmt.Errorf("parsing terragrunt dependency graph: %w", err)
	}
	for path, depPaths := range results {
		// If absolute path then convert to path relative to pug's working
		// directory.
		if filepath.IsAbs(path) {
			var err error
			if path, err = s.stripWorkdirFromPath(path); err != nil {
				s.logger.Error("loading terragrunt dependencies", "error", err)
				// Skip loading dependencies for this module
				continue
			}
		}
		// Retrieve module. If it cannot be found it is probably because the
		// module is outside of pug's working directory, in which case classify
		// it as a warning rather than an error.
		mod, err := s.GetByPath(path)
		if err != nil {
			if errors.Is(err, resource.ErrNotFound) {
				s.logger.Warn("loading terragrunt dependencies", "error", err)
			} else {
				s.logger.Error("loading terragrunt dependencies", "error", err)
			}
			// Skip handling dependencies for this module.
			continue
		}
		// Convert dependency paths to module IDs
		dependencyIDs := make([]resource.ID, 0, len(depPaths))
		for _, path := range depPaths {
			// If absolute path then convert to path relative to pug's working
			// directory.
			if filepath.IsAbs(path) {
				var err error
				if path, err = s.stripWorkdirFromPath(path); err != nil {
					// Skip loading this dependency
					return err
				}
			}
			// Retrieve module. If it cannot be found it is probably because the
			// module is outside of pug's working directory, in which case classify
			// it as a warning rather than an error.
			mod, err := s.GetByPath(path)
			if err != nil {
				if errors.Is(err, resource.ErrNotFound) {
					s.logger.Warn("loading terragrunt dependency", "error", err)
				} else {
					s.logger.Error("loading terragrunt dependency", "error", err)
				}
				// Skip loading this dependency
				continue
			}
			dependencyIDs = append(dependencyIDs, mod.ID)
		}
		s.table.Update(mod.ID, func(existing *Module) error {
			existing.Common = existing.WithDependencies(dependencyIDs...)
			return nil
		})
	}
	return nil
}

func (s *Service) stripWorkdirFromPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Rel(s.workdir.String(), path)
	}
	return path, nil
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
			// Trigger a workspace reload if the module doesn't yet have a
			// current workspace
			if mod.CurrentWorkspaceID == nil {
				s.Publish(resource.UpdatedEvent, mod)
			}
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
	return nil, fmt.Errorf("%s: %w", path, resource.ErrNotFound)
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
	})
}

func (s *Service) Validate(moduleID resource.ID) (*task.Task, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return nil, fmt.Errorf("validating module: %w", err)
	}

	return s.CreateTask(mod, task.CreateOptions{
		Command: []string{"validate"},
	})
}

// TODO: move this logic into task.Create
func (s *Service) CreateTask(mod *Module, opts task.CreateOptions) (*task.Task, error) {
	opts.Parent = mod
	opts.Path = mod.Path
	return s.tasks.Create(opts)
}
