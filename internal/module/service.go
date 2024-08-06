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

type ReloadResult struct {
	// Loaded is true if the module has been newly loaded or false if it has
	// been unloaded
	Loaded bool
	Module *Module
}

// Load searches the working directory recursively for modules and loads them
// into Pug. Newly found modules are sent to the loaded channel. Any modules
// previously loaded into Pug but that can no longer be found are unloaded from
// Pug and sent to the unloaded channel.
func (s *Service) Reload(ctx context.Context) chan ReloadResult {
	ch, errc := find(ctx, s.workdir)
	var results = make(chan ReloadResult, 1)
	go func() {
		var (
			loadTotal int
			// Keep reference to found paths to deduce which modules can no
			// longer be found.
			found []string
		)
		for ch != nil || errc != nil {
			select {
			case opts, ok := <-ch:
				if !ok {
					ch = nil
					break
				}
				found = append(found, opts.Path)
				// Handle found module
				if mod, err := s.GetByPath(opts.Path); errors.Is(err, resource.ErrNotFound) {
					// Not found, so add to pug
					mod := New(s.workdir, opts)
					s.table.Add(mod.ID, mod)
					results <- ReloadResult{Loaded: true, Module: mod}
					loadTotal++
				} else if err != nil {
					s.logger.Error("loading modules", "error", err)
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
					s.logger.Error("loading modules", "error", err)
				}
			}
		}
		s.logger.Info("finished loading modules", "loaded", loadTotal)

		// Cleanup existing modules, removing those that are no longer to be found
		var unloadTotal int
		for _, existing := range s.table.List() {
			if !slices.Contains(found, existing.Path) {
				s.table.Delete(existing.ID)
				unloadTotal++
				results <- ReloadResult{Loaded: false, Module: existing}
			}
		}
		if unloadTotal > 0 {
			s.logger.Info("finished unloading modules", "unloaded", loadTotal)
		}

		if s.terragrunt {
			if err := s.loadTerragruntDependencies(); err != nil {
				s.logger.Error("loading terragrunt dependencies: %w", err)
			}
			// TODO: log info about loaded dependencies
		}
		// Let caller know we're done.
		close(results)
	}()
	return results
}

func (s *Service) loadTerragruntDependencies() error {
	task, err := s.tasks.Create(task.Spec{
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
func (s *Service) Init(moduleID resource.ID) (task.Spec, error) {
	return s.updateSpec(moduleID, task.Spec{
		Command:  []string{"init"},
		Args:     []string{"-input=false"},
		Blocking: true,
		// The terraform plugin cache is not concurrency-safe, so only allow one
		// init task to run at any given time.
		Exclusive: s.pluginCache,
		AfterCreate: func(task *task.Task) {
			// Trigger a workspace reload if the module doesn't yet have a
			// current workspace
			mod := task.Module().(*Module)
			if mod.CurrentWorkspaceID == nil {
				s.Publish(resource.UpdatedEvent, mod)
			}
		},
	})
}

func (s *Service) Format(moduleID resource.ID) (task.Spec, error) {
	return s.updateSpec(moduleID, task.Spec{
		Command: []string{"fmt"},
	})
}

func (s *Service) Validate(moduleID resource.ID) (task.Spec, error) {
	return s.updateSpec(moduleID, task.Spec{
		Command: []string{"validate"},
	})
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

// updateSpec updates the task spec with common module settings.
func (s *Service) updateSpec(moduleID resource.ID, spec task.Spec) (task.Spec, error) {
	mod, err := s.table.Get(moduleID)
	if err != nil {
		return task.Spec{}, err
	}
	spec.Parent = mod
	spec.Path = mod.Path
	return spec, nil
}
