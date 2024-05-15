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

	modules moduleService
	tasks   *task.Service

	*pubsub.Broker[*Workspace]
}

type ServiceOptions struct {
	TaskService   *task.Service
	ModuleService *module.Service
	Logger        logging.Interface
}

type workspaceTable interface {
	Add(id resource.ID, row *Workspace)
	Update(id resource.ID, updater func(existing *Workspace) error) (*Workspace, error)
	Delete(id resource.ID)
	Get(id resource.ID) (*Workspace, error)
	List() []*Workspace
}

type moduleService interface {
	Get(id resource.ID) (*module.Module, error)
	GetByPath(path string) (*module.Module, error)
	SetCurrent(moduleID, workspaceID resource.ID) error
	Reload() ([]string, []string, error)
	List() []*module.Module
	CreateTask(mod *module.Module, opts task.CreateOptions) (*task.Task, error)
}

type moduleSubscription interface {
	Subscribe() <-chan resource.Event[*module.Module]
}

func NewService(opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Workspace](opts.Logger)
	table := resource.NewTable(broker)

	opts.Logger.AddEnricher(&logEnricher{table: table})

	svc := &Service{
		Broker:  broker,
		table:   table,
		modules: opts.ModuleService,
		tasks:   opts.TaskService,
		logger:  opts.Logger,
	}
	return svc
}

// LoadWorkspacesUponModuleLoad automatically loads workspaces for a module
// whenever:
// * a new module is loaded into pug for the first time, unless it is yet to be
// initialized (because `terraform workspace list` would fail)
// * an existing module is updated, has been initialized, and does not yet have
// a current workspace.
func (s *Service) LoadWorkspacesUponModuleLoad(ms moduleSubscription) {
	sub := ms.Subscribe()

	go func() {
		for event := range sub {
			switch event.Type {
			case resource.CreatedEvent:
				if event.Payload.Initialized == nil || *event.Payload.Initialized {
					// Module is new, and has a .terraform dir, so try
					// running `workspace list`
					s.Reload(event.Payload.ID)
				}
			case resource.UpdatedEvent:
				if event.Payload.CurrentWorkspaceID != nil {
					continue
				}
				if event.Payload.Initialized == nil {
					// Initialization status unknown; ignore
					continue
				}
				if !*event.Payload.Initialized {
					// No .terraform dir, or terraform init has failed; ignore
					continue
				}
				s.Reload(event.Payload.ID)
			}
		}
	}()
}

// Reload invokes `terraform workspace list` on a module and updates pug with
// the results, adding any newly discovered workspaces and pruning any
// workspaces no longer found to exist.
func (s *Service) Reload(moduleID resource.ID) (*task.Task, error) {
	mod, err := s.modules.Get(moduleID)
	if err != nil {
		return nil, err
	}
	task, err := s.modules.CreateTask(mod, task.CreateOptions{
		Command: []string{"workspace", "list"},
		AfterError: func(t *task.Task) {
			s.logger.Error("reloading workspaces", "error", t.Err, "module", mod, "task", t)
		},
		AfterExited: func(t *task.Task) {
			found, current, err := parseList(t.NewReader())
			if err != nil {
				s.logger.Error("reloading workspaces", "error", err, "module", mod, "task", t)
				return
			}
			added, removed, err := s.resetWorkspaces(mod, found, current)
			if err != nil {
				s.logger.Error("reloading workspaces", "error", err, "module", mod, "task", t)
				return
			}
			s.logger.Info("reloaded workspaces", "added", added, "removed", removed, "module", mod)
		},
	})
	if err != nil {
		s.logger.Error("reloading workspaces", "error", err, "module", mod)
		return nil, err
	}
	return task, nil
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
func parseList(r io.Reader) (list []string, current string, err error) {
	// Reader should output something like this:
	//
	//   default
	//   non-default-1
	// * non-default-2
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		out := strings.TrimSpace(scanner.Text())
		if out == "" {
			continue
		}
		// Handle current workspace denoted with a "*" prefix
		if strings.HasPrefix(out, "*") {
			var found bool
			_, current, found = strings.Cut(out, "* ")
			if !found {
				return nil, "", fmt.Errorf("malformed output: %s", out)
			}
			out = current
		}
		list = append(list, out)
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return
}

// Create a workspace. Asynchronous.
func (s *Service) Create(path, name string) (*Workspace, *task.Task, error) {
	mod, err := s.modules.GetByPath(path)
	if err != nil {
		return nil, nil, fmt.Errorf("checking for module: %s: %w", path, err)
	}
	ws, err := New(mod, name)
	if err != nil {
		return nil, nil, err
	}

	task, err := s.createTask(ws, task.CreateOptions{
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
	})
	if err != nil {
		return nil, nil, err
	}
	return ws, task, nil
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
	// Create task to immediately set workspace as current workspace for module.
	task, err := s.createTask(ws, task.CreateOptions{
		Command:   []string{"workspace", "select"},
		Args:      []string{ws.Name},
		Immediate: true,
	})
	if err != nil {
		return err
	}
	if err := task.Wait(); err != nil {
		return err
	}
	// Now task has finished successfully, update the current workspace in pug
	// as well.
	if err := s.modules.SetCurrent(moduleID, workspaceID); err != nil {
		return err
	}
	return nil
}

// Delete a workspace. Asynchronous.
func (s *Service) Delete(id resource.ID) (*task.Task, error) {
	ws, err := s.table.Get(id)
	if err != nil {
		return nil, fmt.Errorf("deleting workspace: %w", err)
	}
	return s.createTask(ws, task.CreateOptions{
		Command:  []string{"workspace", "delete"},
		Args:     []string{ws.Name},
		Blocking: true,
		AfterExited: func(*task.Task) {
			s.table.Delete(ws.ID)
		},
	})
}

// TODO: move this logic into task.Create
func (s *Service) createTask(ws *Workspace, opts task.CreateOptions) (*task.Task, error) {
	opts.Parent = ws

	mod, err := s.modules.Get(ws.ModuleID())
	if err != nil {
		return nil, err
	}
	opts.Path = mod.Path

	return s.tasks.Create(opts)
}
