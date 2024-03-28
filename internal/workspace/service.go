package workspace

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	broker *pubsub.Broker[*Workspace]
	table  workspaceTable

	modules moduleService
	tasks   *task.Service
}

type ServiceOptions struct {
	TaskService   *task.Service
	ModuleService *module.Service
}

type workspaceTable interface {
	Add(id resource.ID, row *Workspace)
	Update(id resource.ID, updater func(existing *Workspace)) error
	Delete(id resource.ID)
	Get(id resource.ID) (*Workspace, error)
	List() []*Workspace
}

type moduleService interface {
	Get(id resource.ID) (*module.Module, error)
	GetByPath(path string) (*module.Module, error)
	SetCurrent(moduleID resource.ID, workspace resource.Resource) error
	Reload() error
	List() []*module.Module
}

func NewService(ctx context.Context, opts ServiceOptions) *Service {
	broker := pubsub.NewBroker[*Workspace]()
	svc := &Service{
		broker:  broker,
		table:   resource.NewTable(broker),
		modules: opts.ModuleService,
		tasks:   opts.TaskService,
	}
	return svc
}

// Reload invokes `terraform workspace list` on a module and updates pug with
// the results, adding any newly discovered workspaces and pruning any
// workspaces no longer found to exist.
//
// TODO: only permit one reload at a time.
func (s *Service) Reload(moduleID resource.ID) (*task.Task, error) {
	mod, err := s.modules.Get(moduleID)
	if err != nil {
		return nil, err
	}
	task, err := s.tasks.Create(task.CreateOptions{
		Parent:  mod.Resource,
		Path:    moduleID.String(),
		Command: []string{"workspace", "list"},
		AfterError: func(t *task.Task) {
			slog.Error("reloading workspaces", "error", t.Err, "module", moduleID, "task", t)
		},
		AfterExited: func(t *task.Task) {
			found, current, err := parseList(t.NewReader())
			if err != nil {
				slog.Error("reloading workspaces", "error", err, "module", moduleID)
				return
			}
			added, removed, err := s.resetWorkspaces(mod.Resource, found, current)
			if err != nil {
				slog.Error("reloading workspaces", "error", err, "module", moduleID)
				return
			}
			slog.Info("reloaded workspaces", "added", added, "removed", removed, "module", moduleID)
		},
	})
	if err != nil {
		slog.Error("reloading workspaces", "error", err, "module", moduleID)
		return nil, err
	}
	slog.Info("reloading workspaces", "module", moduleID, "task", task)
	return task, nil
}

func (s *Service) ReloadAll() (task.Multi, []error) {
	if err := s.modules.Reload(); err != nil {
		return nil, []error{err}
	}
	mods := s.modules.List()

	modIDs := make([]resource.ID, len(mods))
	for i, mod := range s.modules.List() {
		modIDs[i] = mod.ID()
	}
	return task.CreateMulti(s.Reload, modIDs...)
}

// resetWorkspaces resets the workspaces for a module, adding newly discovered
// workspaces, removing workspaces that no longer exist, and setting the current
// workspace for the module.
func (s *Service) resetWorkspaces(module resource.Resource, discovered []string, current string) (added []string, removed []string, err error) {
	// Gather existing workspaces for the module.
	var existing []*Workspace
	for _, ws := range s.table.List() {
		if ws.Module().ID() == module.ID() {
			existing = append(existing, ws)
		}
	}

	// Add discovered workspaces that don't exist in pug
	for _, name := range discovered {
		if !slices.ContainsFunc(existing, func(ws *Workspace) bool {
			return ws.String() == name
		}) {
			add := New(module, name)
			s.table.Add(add.ID(), add)
			added = append(added, name)
		}
	}
	// Remove workspaces from pug that no longer exist
	for _, ws := range existing {
		if !slices.Contains(discovered, ws.String()) {
			s.table.Delete(ws.ID())
			removed = append(removed, ws.String())
		}
	}
	// Reset current workspace
	currentWorkspace, err := s.GetByName(module.ID(), current)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot find current workspace: %s: %w", current, err)
	}
	if err := s.modules.SetCurrent(module.ID(), currentWorkspace.Resource); err != nil {
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
	ws := New(mod.Resource, name)

	task, err := s.createTask(ws, task.CreateOptions{
		Command: []string{"workspace", "new"},
		Args:    []string{name},
		AfterExited: func(*task.Task) {
			s.table.Add(ws.ID(), ws)
			// `workspace new` implicitly makes the created workspace the
			// *current* workspace, so better tell pug that too.
			if err := s.modules.SetCurrent(mod.ID(), ws.Resource); err != nil {
				slog.Error("creating workspace: %w", err)
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
		if ws.Module().ID() == moduleID && ws.Name() == name {
			return ws, nil
		}
	}
	return nil, resource.ErrNotFound
}

type ListOptions struct {
	ModuleID resource.ID
}

func (s *Service) List(opts ListOptions) []*Workspace {
	var existing []*Workspace
	for _, ws := range s.table.List() {
		if opts.ModuleID != resource.GlobalID {
			if ws.Module().ID() != opts.ModuleID {
				continue
			}
		}
		existing = append(existing, ws)
	}
	return existing
}

func (s *Service) Subscribe(ctx context.Context) <-chan resource.Event[*Workspace] {
	return s.broker.Subscribe(ctx)
}

func (s *Service) SetCurrent(workspaceID resource.ID, run resource.Resource) {
	ws, err := s.table.Get(workspaceID)
	if err != nil {
		slog.Error("setting current workspace run", "run", run, "error", err)
	}
	ws.CurrentRun = &run
}

// Delete a workspace. Asynchronous.
func (s *Service) Delete(id resource.ID) (*task.Task, error) {
	ws, err := s.table.Get(id)
	if err != nil {
		return nil, fmt.Errorf("deleting workspace: %w", err)
	}

	return s.createTask(ws, task.CreateOptions{
		Command:  []string{"workspace", "delete"},
		Args:     []string{ws.String()},
		Blocking: true,
		AfterExited: func(*task.Task) {
			s.table.Delete(ws.ID())
		},
	})
}

func (s *Service) createTask(ws *Workspace, opts task.CreateOptions) (*task.Task, error) {
	opts.Parent = ws.Resource
	opts.Path = ws.ModulePath()
	return s.tasks.Create(opts)
}
