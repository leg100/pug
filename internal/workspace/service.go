package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	table  workspaceTable
	logger logging.Interface

	modules modules
	tasks   *task.Service

	*pubsub.Broker[*Workspace]
	*reloader
}

type ServiceOptions struct {
	Tasks   *task.Service
	Modules *module.Service
	Logger  logging.Interface
}

type workspaceTable interface {
	Add(id resource.ID, row *Workspace)
	Update(id resource.ID, updater func(existing *Workspace) error) (*Workspace, error)
	Delete(id resource.ID)
	Get(id resource.ID) (*Workspace, error)
	List() []*Workspace
}

type modules interface {
	Get(id resource.ID) (*module.Module, error)
	GetByPath(path string) (*module.Module, error)
	SetCurrent(moduleID, workspaceID resource.ID) error
	SetLoadedWorkspaces(moduleID resource.ID) error
	Reload() ([]string, []string, error)
	List() []*module.Module
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Workspace](opts.Logger)
	table := resource.NewTable(broker)

	opts.Logger.AddEnricher(&logEnricher{table: table})

	s := &Service{
		Broker:  broker,
		table:   table,
		modules: opts.Modules,
		tasks:   opts.Tasks,
		logger:  opts.Logger,
	}
	s.reloader = &reloader{s}
	return s
}

// LoadWorkspaces is called by the module constructor to load its
// initial workspaces including a default workspace as well its
// current workspace, the ID of which is returned.
func (s *Service) LoadWorkspaces(mod *module.Module) (resource.ID, error) {
	// Load default workspace
	defaultWorkspace, err := New(mod, "default")
	if err != nil {
		return resource.ID{}, fmt.Errorf("loading default workspace: %w", err)
	}
	s.table.Add(defaultWorkspace.ID, defaultWorkspace)

	// Determine current workspace. If a .terraform/environment file exists then
	// read current workspace from there.
	envfile, err := os.ReadFile(filepath.Join(mod.FullPath(), ".terraform", "environment"))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			s.logger.Error("reading current workspace from file", "error", err)
		}
		// Current workspace is the default workspace if there is any error
		// reading the environment file.
		return defaultWorkspace.ID, nil
	}
	current := string(bytes.TrimSpace(envfile))
	if current == "default" {
		// Nothing more to be done.
		return defaultWorkspace.ID, nil
	}

	// Load non-default workspace and return it as the current worskpace.
	nonDefaultWorkspace, err := New(mod, current)
	if err != nil {
		return resource.ID{}, fmt.Errorf("loading current workspace: %w", err)
	}
	s.table.Add(nonDefaultWorkspace.ID, nonDefaultWorkspace)
	return nonDefaultWorkspace.ID, nil
}

// LoadWorkspacesUponModuleLoad automatically loads workspaces for a module
// that has been newly loaded into pug.
func (s *Service) LoadWorkspacesUponModuleLoad(sub <-chan resource.Event[*module.Module]) {
	for event := range sub {
		if event.Type != resource.CreatedEvent {
			continue
		}
		// Should be false on a new module.
		if event.Payload.LoadedWorkspaces {
			continue
		}
		if err := s.createReloadTask(event.Payload.ID); err != nil {
			s.logger.Error("reloading workspaces", "module", event.Payload)
		}
	}
}

// LoadWorkspacesUponInit automatically loads workspaces for a module whenever
// it is successfully initialized and the module has not yet had its workspaces
// loaded.
func (s *Service) LoadWorkspacesUponInit(sub <-chan resource.Event[*task.Task]) {
	for event := range sub {
		if !module.IsInitTask(event.Payload) {
			continue
		}
		if event.Payload.State != task.Exited {
			continue
		}
		mod, err := s.modules.Get(event.Payload.Module().GetID())
		if err != nil {
			continue
		}
		if mod.LoadedWorkspaces {
			continue
		}
		if err := s.createReloadTask(mod.ID); err != nil {
			s.logger.Error("reloading workspaces", "module", event.Payload)
		}
	}
}

// Create a workspace. Asynchronous.
func (s *Service) Create(path, name string) (task.Spec, error) {
	mod, err := s.modules.GetByPath(path)
	if err != nil {
		return task.Spec{}, err
	}
	ws, err := New(mod, name)
	if err != nil {
		return task.Spec{}, err
	}
	return task.Spec{
		Parent:  mod,
		Path:    mod.Path,
		Command: []string{"workspace", "new"},
		Args:    []string{name},
		AfterExited: func(*task.Task) {
			s.table.Add(ws.ID, ws)
			// `workspace new` implicitly makes the created workspace the
			// *current* workspace, so better tell pug that too.
			if err := s.modules.SetCurrent(mod.ID, ws.ID); err != nil {
				s.logger.Error("creating workspace: %w", err)
			}
		},
	}, nil
}

func (s *Service) Get(workspaceID resource.ID) (*Workspace, error) {
	return s.table.Get(workspaceID)
}

func (s *Service) GetByName(moduleID resource.ID, name string) (*Workspace, error) {
	for _, ws := range s.table.List() {
		if ws.ModuleID() == moduleID && ws.Name == name {
			return ws, nil
		}
	}
	return nil, resource.ErrNotFound
}

type ListOptions struct {
	// Filter by ID of workspace's module. If zero value then no filtering is
	// performed.
	ModuleID resource.ID
}

func (s *Service) List(opts ListOptions) []*Workspace {
	var existing []*Workspace
	for _, ws := range s.table.List() {
		// If opts.ModuleID is zero value then HasAncestor returns true.
		if !ws.HasAncestor(opts.ModuleID) {
			continue
		}
		existing = append(existing, ws)
	}
	return existing
}

// SelectWorkspace runs the `terraform workspace select <workspace_name>`
// command, which sets the current workspace for the module. Once that's
// finished it then updates the current workspace in pug itself too.
func (s *Service) SelectWorkspace(moduleID, workspaceID resource.ID) error {
	ws, err := s.table.Get(workspaceID)
	if err != nil {
		return err
	}
	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return err
	}
	// Create task to immediately set workspace as current workspace for module.
	_, err = s.tasks.Create(task.Spec{
		Parent:    mod,
		Path:      mod.Path,
		Command:   []string{"workspace", "select"},
		Args:      []string{ws.Name},
		Immediate: true,
		Wait:      true,
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			// Now the terraform command has finished, update the current
			// workspace in pug as well.
			err := s.modules.SetCurrent(moduleID, workspaceID)
			return nil, err
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete a workspace. Asynchronous.
func (s *Service) Delete(workspaceID resource.ID) (task.Spec, error) {
	ws, err := s.table.Get(workspaceID)
	if err != nil {
		return task.Spec{}, fmt.Errorf("deleting workspace: %w", err)
	}
	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return task.Spec{}, err
	}
	return task.Spec{
		Parent:   mod,
		Path:     mod.Path,
		Command:  []string{"workspace", "delete"},
		Args:     []string{ws.Name},
		Blocking: true,
		AfterExited: func(*task.Task) {
			s.table.Delete(ws.ID)
		},
	}, nil
}
