package workspace

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

type Service struct {
	modules    *module.Service
	tasks      *task.Service
	broker     *pubsub.Broker[*Workspace]
	workspaces map[uuid.UUID]*Workspace
	// Mutex for concurrent read/write of workspaces
	mu sync.Mutex
}

// Reload the store with a fresh list of workspaces discovered by running
// `terraform workspace list` in each module. Any workspaces currently stored
// but no longer found are pruned.
func (s *Service) Reload() ([]*task.Task, error) {
	mods := s.modules.List()
	tasks := make([]*task.Task, len(mods))
	for i, m := range mods {
		t, err := s.modules.CreateTask(m.ID, task.CreateOptions{
			Command: []string{"workspace", "list"},
			AfterError: func(t *task.Task) {
				// TODO: log error and prune workspaces
			},
			AfterExited: func(t *task.Task) {
				found, current, err := parseList(t.NewReader())
				if err != nil {
					slog.Error("reloading workspaces", "error", err, "module", m)
					return
				}
				s.resetWorkspaces(m.Path, found)
				// prune existing workspaces that were not found in module.
				for _, ws := range s.workspaces {
					if ws.Module().ID == m.ID {
						if !slices.ContainsFunc(found, func(foundws *Workspace) bool {
							return foundws.Name == ws.Name
						}) {
							s.mu.Lock()
							delete(s.workspaces, ws.ID)
							s.mu.Unlock()
						}
					}
				}
			},
		})
		if err != nil {
			slog.Error("reloading workspaces", "error", err, "module", m)
			return nil, err
		}
		tasks[i] = t
	}
	// TODO: log created tasks
	return tasks, nil
}

// reload workspaces for a module:
// parameters:
// 1) discovered names of workspaces to be added, including whether it is the current
// workspace for the module.
// 2) existing workspaces
// return:
// 1) jjjj

// reloadModule reloads the workspaces for a particular module.
func (s *Service) resetWorkspaces(module *Module, discovered []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.ListByModule(moduleID)
	if err != nil {
		return err
	}
	// Add discovered workspaces that don't exist yet
	for _, name := range discovered {
		if !slices.ContainsFunc(existing, func(ws *Workspace) bool {
			return ws.Name == name
		}) {
			ws := newWorkspace(module, name, 
		}
	}
	// 
	for _, ws := range existing {
		if !slices.Contains(discovered, ws.Name) {
			delete(s.workspaces, ws.ID)
		}
	}
}

// resetCurrent resets the current workspace for a module.
func (s *Service) resetCurrent(module uuid.UUID, current string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ws := range s.workspaces {
		if ws.Module().ID == module {
			ws.Current = (ws.Name == current)
		}
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
			_, current, found = strings.Cut(out, " ")
			if !found {
				return nil, "", fmt.Errorf("malformed output: %s", out)
			}
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
	ws := &Workspace{
		Resource: resource.New(&mod.Resource),
		Name:     name,
	}

	task, err := s.modules.CreateTask(mod.ID, task.CreateOptions{
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
func (s *Service) Get(id uuid.UUID) (*Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return ws, nil
}

// List workspaces by module.
func (s *Service) ListByModule(moduleID uuid.UUID) ([]*Workspace, error) {
	mod, err := s.modules.Get(moduleID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var workspaces []*Workspace
	for _, ws := range s.workspaces {
		if ws.Module().ID == mod.ID {
			workspaces = append(workspaces, ws)
		}
	}
	return workspaces, nil
}

// Delete a workspace. Asynchronous.
func (s *Service) Delete(id uuid.UUID) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id]
	if !ok {
		return nil, resource.ErrNotFound
	}

	return s.modules.CreateTask(ws.Parent.ID, task.CreateOptions{
		Command:  []string{"workspace", "delete"},
		Args:     []string{ws.Name},
		Blocking: true,
		AfterExited: func(*task.Task) {
			s.mu.Lock()
			delete(s.workspaces, id)
			s.mu.Unlock()

			s.broker.Publish(resource.DeletedEvent, ws)
		},
	})
}

const MetadataKey = "workspace"

// Create workspace task.
func (s *Service) CreateTask(id uuid.UUID, opts task.CreateOptions) (*task.Task, error) {
	// ensure workspace exists
	ws, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %s: %w", id, err)
	}
	opts.Env = append(opts.Env, fmt.Sprintf("TF_WORKSPACE=%s", ws.Name))
	return s.modules.CreateTask(ws.Parent.ID, opts)
}
