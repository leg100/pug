package workspace

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	modules    *module.Service
	tasks      *task.Service
	broker     *pubsub.Broker[*Workspace]
	workspaces map[resource.ID]*Workspace
	// Mutex for concurrent read/write of workspaces
	mu sync.Mutex
}

type ServiceOptions struct {
	TaskService   *task.Service
	ModuleService *module.Service
}

func NewService(ctx context.Context, opts ServiceOptions) *Service {
	svc := &Service{
		modules:    opts.ModuleService,
		tasks:      opts.TaskService,
		broker:     pubsub.NewBroker[*Workspace](),
		workspaces: make(map[resource.ID]*Workspace),
	}
	// Load workspaces whenever a module is created.
	sub, _ := opts.ModuleService.Subscribe(ctx)
	go func() {
		for event := range sub {
			switch event.Type {
			case resource.CreatedEvent:
				_, _ = svc.Reload(event.Payload.Resource)
			}
		}
	}()
	return svc
}

// Reload invokes `terraform workspace list` on a module and updates pug with
// the results, adding any newly discovered workspaces and pruning any
// workspaces no longer found to exist.
func (s *Service) Reload(module resource.Resource) (*task.Task, error) {
	// TODO: only permit one reload at a time.

	task, err := s.tasks.Create(task.CreateOptions{
		Parent:  module,
		Path:    module.String(),
		Command: []string{"workspace", "list"},
		AfterError: func(t *task.Task) {
			// TODO: log error and prune workspaces
		},
		AfterExited: func(t *task.Task) {
			found, current, err := parseList(t.NewReader())
			if err != nil {
				slog.Error("reloading workspaces", "error", err, "module", module)
				return
			}
			s.resetWorkspaces(module, found, current)
		},
	})
	if err != nil {
		slog.Error("reloading workspaces", "error", err, "module", module)
		return nil, err
	}
	return task, nil
}

// resetWorkspaces resets the workspaces for a module, adding newly discovered
// workspaces, removing workspaces that no longer exist, and setting the current
// workspace for the module.
func (s *Service) resetWorkspaces(module resource.Resource, discovered []string, current string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Gather existing workspaces for the module.
	var existing []*Workspace
	for _, ws := range s.workspaces {
		if ws.Module().ID == module.ID {
			existing = append(existing, ws)
		}
	}

	// Add discovered workspaces that don't exist in pug
	for _, name := range discovered {
		if !slices.ContainsFunc(existing, func(ws *Workspace) bool {
			return ws.String() == name
		}) {
			add := newWorkspace(module, name, false)
			s.workspaces[add.ID] = add
			s.broker.Publish(resource.CreatedEvent, add)
		}
	}
	// Remove workspaces from pug that no longer exist
	for _, ws := range existing {
		if !slices.Contains(discovered, ws.String()) {
			delete(s.workspaces, ws.ID)
			s.broker.Publish(resource.DeletedEvent, ws)
		}
	}
	// Reset current workspace
	for _, ws := range s.workspaces {
		ws.Current = (ws.String() == current)
	}
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
	// `terraform workspace new` below implicitly makes the created workspace
	// the *current* workspace.
	ws := newWorkspace(mod.Resource, name, true)

	task, err := s.createTask(ws, task.CreateOptions{
		Command: []string{"workspace", "new"},
		Args:    []string{name},
		AfterExited: func(*task.Task) {
			s.mu.Lock()
			s.workspaces[ws.ID] = ws
			s.mu.Unlock()

			s.broker.Publish(resource.CreatedEvent, ws)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return ws, task, nil
}

// Get a workspace.
func (s *Service) Get(id resource.ID) (*Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return ws, nil
}

// Get the current workspace for a module
func (s *Service) GetCurrent(moduleID resource.ID) (*Workspace, error) {
	moduleWorkspaces := s.List(ListOptions{
		ModuleID: &moduleID,
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ws := range moduleWorkspaces {
		if ws.Current {
			return ws, nil
		}
	}
	// Should never happen.
	return nil, resource.ErrNotFound
}

type ListOptions struct {
	ModuleID *resource.ID
}

func (s *Service) List(opts ListOptions) []*Workspace {
	s.mu.Lock()
	defer s.mu.Unlock()

	var existing []*Workspace
	for _, ws := range s.workspaces {
		if opts.ModuleID != nil {
			if ws.Module().ID != *opts.ModuleID {
				continue
			}
		}
		existing = append(existing, ws)
	}
	return existing
}

func (s *Service) Subscribe(ctx context.Context) (<-chan resource.Event[*Workspace], func()) {
	return s.broker.Subscribe(ctx)
}

// Delete a workspace. Asynchronous.
func (s *Service) Delete(id resource.ID) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id]
	if !ok {
		return nil, resource.ErrNotFound
	}

	return s.createTask(ws, task.CreateOptions{
		Command:  []string{"workspace", "delete"},
		Args:     []string{ws.String()},
		Blocking: true,
		AfterExited: func(*task.Task) {
			s.mu.Lock()
			delete(s.workspaces, id)
			s.mu.Unlock()

			s.broker.Publish(resource.DeletedEvent, ws)
		},
	})
}

func (s *Service) createTask(ws *Workspace, opts task.CreateOptions) (*task.Task, error) {
	mod, err := s.modules.Get(ws.Module().ID)
	if err != nil {
		return nil, err
	}
	opts.Parent = ws.Resource
	opts.Path = mod.Path
	return s.tasks.Create(opts)
}
