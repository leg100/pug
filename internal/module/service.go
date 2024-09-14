package module

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"

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
	Create(spec task.Spec) (*task.Task, error)
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

	opts.Logger.AddArgsUpdater(&logging.ReferenceUpdater[*Module]{
		Getter: table,
		Name:   "module",
		Field:  "ModuleID",
	})

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
//
// TODO: separate into Load and Reload
func (s *Service) Reload() (added []string, removed []string, err error) {
	ch, errc := find(context.TODO(), s.workdir)
	var found []string
	for ch != nil || errc != nil {
		select {
		case opts, ok := <-ch:
			if !ok {
				ch = nil
				break
			}
			found = append(found, opts.Path)
			// handle found module
			if mod, err := s.GetByPath(opts.Path); errors.Is(err, resource.ErrNotFound) {
				// Not found, so add to pug
				mod := New(opts)
				s.table.Add(mod.ID, mod)
				added = append(added, opts.Path)
			} else if err != nil {
				s.logger.Error("reloading modules", "error", err)
			} else {
				// Update in-place; the backend may have changed.
				s.table.Update(mod.ID, func(existing *Module) error {
					existing.Backend = opts.Backend
					return nil
				})
			}
		case err, ok := <-errc:
			if !ok {
				errc = nil
				break
			}
			if err != nil {
				s.logger.Error("reloading modules", "error", err)
			}
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

	if s.terragrunt {
		if err := s.loadTerragruntDependencies(); err != nil {
			s.logger.Error("loading terragrunt dependencies: %w", err)
		}
	}
	return
}

func (s *Service) loadTerragruntDependencies() error {
	task, err := s.tasks.Create(task.Spec{
		Execution: task.Execution{
			TerraformCommand: []string{"graph-dependencies"},
		},
		Wait: true,
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
			if path, err = s.workdir.Rel(path); err != nil {
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
				if path, err = s.workdir.Rel(path); err != nil {
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
			existing.dependencies = dependencyIDs
			return nil
		})
	}
	return nil
}

const InitTask task.Identifier = "init"

// Init invokes terraform init on the module.
func (s *Service) Init(moduleID resource.ID, upgrade bool) (task.Spec, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return task.Spec{}, err
	}
	args := []string{"-input=false"}
	if upgrade {
		args = append(args, "-upgrade")
	}
	spec := task.Spec{
		ModuleID:   &mod.ID,
		Path:       mod.Path,
		Identifier: InitTask,
		Execution: task.Execution{
			TerraformCommand: []string{"init"},
			Args:             args,
		},
		Blocking: true,
		// The terraform plugin cache is not concurrency-safe, so only allow one
		// init task to run at any given time.
		Exclusive: s.pluginCache,
	}
	return spec, nil
}

func (s *Service) Format(moduleID resource.ID) (task.Spec, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return task.Spec{}, err
	}
	spec := task.Spec{
		ModuleID: &mod.ID,
		Path:     mod.Path,
		Execution: task.Execution{
			TerraformCommand: []string{"fmt"},
		},
	}
	return spec, nil
}

func (s *Service) Validate(moduleID resource.ID) (task.Spec, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return task.Spec{}, err
	}
	spec := task.Spec{
		ModuleID: &mod.ID,
		Path:     mod.Path,
		Execution: task.Execution{
			TerraformCommand: []string{"validate"},
		},
	}
	return spec, nil
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

// SetCurrent sets the current workspace for the module.
func (s *Service) SetCurrent(moduleID, workspaceID resource.ID) error {
	_, err := s.table.Update(moduleID, func(existing *Module) error {
		existing.CurrentWorkspaceID = &workspaceID
		return nil
	})
	return err
}

// Execute a program in a module's directory.
func (s *Service) Execute(moduleID resource.ID, program string, args ...string) (task.Spec, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return task.Spec{}, err
	}
	spec := task.Spec{
		ModuleID: &mod.ID,
		Path:     mod.Path,
		Execution: task.Execution{
			Program: program,
			Args:    args,
		},
		// We're executing an arbitrary program which could be performing
		// mutually exclusive actions that prevent other tasks from running as
		// expected, so we make it a blocking task to be on the safe side.
		Blocking: true,
	}
	return spec, nil
}
