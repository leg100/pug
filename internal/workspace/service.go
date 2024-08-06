package workspace

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"

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
	Reload() ([]string, []string, error)
	List() []*module.Module
}

type moduleSubscription interface {
	Subscribe() <-chan resource.Event[*module.Module]
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Workspace](opts.Logger)
	table := resource.NewTable(broker)

	opts.Logger.AddEnricher(&logEnricher{table: table})

	return &Service{
		Broker:  broker,
		table:   table,
		modules: opts.Modules,
		tasks:   opts.Tasks,
		logger:  opts.Logger,
	}
}

// LoadWorkspacesUponModuleLoad automatically loads workspaces for a module
// whenever:
// * a new module is loaded into pug for the first time
// * an existing module is updated and does not yet have a current workspace.
//
// TODO: "load" is ambiguous, it often means the opposite of save, i.e. read
// from a system, whereas what is intended is to save or add workspaces to pug.
func (s *Service) LoadWorkspacesUponModuleLoad(modules moduleSubscription) {
	sub := modules.Subscribe()

	reload := func(moduleID resource.ID) error {
		spec, err := s.Reload(moduleID)
		if err != nil {
			return err
		}
		_, err = s.tasks.Create(spec)
		return err
	}

	go func() {
		for event := range sub {
			switch event.Type {
			case resource.CreatedEvent:
				if err := reload(event.Payload.ID); err != nil {
					s.logger.Error("reloading workspaces", "module", event.Payload)
				}
			case resource.UpdatedEvent:
				if event.Payload.CurrentWorkspaceID != nil {
					// Module already has a current workspace; no need to reload
					// workspaces
					continue
				}
				if err := reload(event.Payload.ID); err != nil {
					s.logger.Error("reloading workspaces", "module", event.Payload)
				}
			}
		}
	}()
}

// Reload returns a task spec that runs `terraform workspace list` on a
// module and updates pug with the results, adding any newly discovered
// workspaces and pruning any workspaces no longer found to exist.
func (s *Service) Reload(moduleID resource.ID) (task.Spec, error) {
	mod, err := s.modules.Get(moduleID)
	if err != nil {
		return task.Spec{}, err
	}
	return task.Spec{
		Parent:      mod,
		Path:        mod.Path,
		Command:     []string{"workspace", "list"},
		Description: "reload workspaces",
		AfterError: func(t *task.Task) {
			s.logger.Error("reloading workspaces", "error", t.Err, "module", mod, "task", t)
		},
		AfterCLISuccess: func(t *task.Task) error {
			found, current, err := parseList(t.NewReader(false))
			if err != nil {
				return err
			}
			added, removed, err := s.resetWorkspaces(mod, found, current)
			if err != nil {
				return err
			}
			s.logger.Info("reloaded workspaces", "added", added, "removed", removed, "module", mod)
			//
			// TODO: write message to stdout
			return nil
		},
	}, nil
}

// resetWorkspaces resets the workspaces for a module, adding newly discovered
// workspaces, removing workspaces that no longer exist, and setting the current
// workspace for the module.
func (s *Service) resetWorkspaces(mod *module.Module, discovered []string, current string) (added []string, removed []string, err error) {
	// Gather existing workspaces for the module.
	var existing []*Workspace
	for _, ws := range s.table.List() {
		if ws.ModuleID() == mod.ID {
			existing = append(existing, ws)
		}
	}

	// Add discovered workspaces that don't exist in pug
	for _, name := range discovered {
		if !slices.ContainsFunc(existing, func(ws *Workspace) bool {
			return ws.Name == name
		}) {
			add, err := New(mod, name)
			if err != nil {
				return nil, nil, fmt.Errorf("adding workspace: %w", err)
			}
			s.table.Add(add.ID, add)
			added = append(added, name)
		}
	}
	// Remove workspaces from pug that no longer exist
	for _, ws := range existing {
		if !slices.Contains(discovered, ws.Name) {
			s.table.Delete(ws.ID)
			removed = append(removed, ws.Name)
		}
	}
	// Reset current workspace
	currentWorkspace, err := s.GetByName(mod.ID, current)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot find current workspace: %s: %w", current, err)
	}
	err = s.modules.SetCurrent(mod.ID, currentWorkspace.ID)
	if err != nil {
		return nil, nil, err
	}
	return
}

// Parse workspaces from the output of `terraform workspace list`.
//
// The output should contain something like this:
//
//	<asterisk> default
//	  non-default-1
//	  non-default-2
func parseList(r io.Reader) (list []string, current string, err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if _, name, found := strings.Cut(scanner.Text(), "* "); found {
			// An asterisk prefix means this is the current workspace.
			current = name
			list = append(list, name)
		} else if _, name, found := strings.Cut(scanner.Text(), "  "); found {
			list = append(list, name)
		} else {
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return
}

// Discovery should work like this. Modules are detected and added to pug as
// soon as they are detected, and when they are added, a task is invoked to list
// its workspaces and load them into pug as well. If the task fails that is ok.
// Discovery should wait for both the module and workspace loading to complete
// before returning.
//
// This could be achieved with channels?
// detect-modules -> add-modules -> detect-workspaces -> add-workspaces
//
// Make this work for:
//
// 1 TUI bootstrap
// 2 CLI bootstrap
// 3 TUI reloading of modules (which can trigger #4) - adds/removes modules
// 4 TUI reloading of workspaces for a module
// 	a. detect-workspaces (terraform workspace list)
// 	b. add/remove workspaces

// Discover populates Pug with modules and workspaces. Synchronous.
func (s *Service) Discover() error {
	// TODO: log discovered modules
	if _, _, err := s.modules.Reload(); err != nil {
		return err
	}
	for _, mod := range s.modules.List() {
		spec, err := s.Reload(mod.ID)
		if err != nil {
			return err
		}
		if _, err = s.tasks.Create(spec); err != nil {
			return err
		}
	}
	return nil
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
		AfterCLISuccess: func(*task.Task) error {
			s.table.Add(ws.ID, ws)
			// `workspace new` implicitly makes the created workspace the
			// *current* workspace, so better tell pug that too.
			//
			// TODO: write message to stdout
			return s.modules.SetCurrent(mod.ID, ws.ID)
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

func (s *Service) SetCurrentRun(workspaceID, runID resource.ID) error {
	ws, err := s.table.Update(workspaceID, func(existing *Workspace) error {
		existing.CurrentRunID = &runID
		return nil
	})
	if err != nil {
		s.logger.Error("setting current run", "workspace_id", workspaceID, "run_id", runID, "error", err)
		return err
	}
	s.logger.Debug("set current run", "workspace", ws, "run_id", runID, "error", err)
	return nil
}

// SelectWorkspace runs the `terraform workspace select <workspace_name>`
// command, which sets the current workspace for the module. Once that's
// finished it then updates the current workspace in pug itself too.
func (s *Service) SelectWorkspace(moduleID, workspaceID resource.ID) error {
	if err := s.selectWorkspace(moduleID, workspaceID); err != nil {
		s.logger.Error("selecting current workspace", "workspace_id", workspaceID, "error", err)
		return err
	}
	s.logger.Debug("selected current workspace", "workspace", workspaceID)
	return nil
}

func (s *Service) selectWorkspace(moduleID, workspaceID resource.ID) error {
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
		AfterCLISuccess: func(*task.Task) error {
			// Now task has finished successfully, update the current workspace in pug
			// as well.
			return s.modules.SetCurrent(moduleID, workspaceID)
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
