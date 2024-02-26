package workspace

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
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
	workspaces map[ID]*Workspace
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
		t, err := s.modules.CreateTask(m.Path, task.CreateOptions{
			Command: []string{"workspace", "list"},
			AfterError: func(t *task.Task) {
				// TODO: log error and prune workspaces
			},
			AfterExited: func(t *task.Task) {
				workspaces, err := parseList(t.Path, t.NewReader())
				if err != nil {
					slog.Error("reloading workspaces", "error", err, "module", m)
					return
				}
				for _, ws := range workspaces {
					s.mu.Lock()
					s.workspaces[ws.Resource] = ws
					s.mu.Unlock()
				}
				// TODO: prune workspaces no longer to be found in module
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

// Parse workspaces from the output of `terraform workspace list`.
func parseList(path string, r io.Reader) ([]*Workspace, error) {
	// Reader should output something like this:
	//
	//   default
	//   non-default-1
	// * non-default-2
	var workspaces []*Workspace
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		out := strings.TrimSpace(scanner.Text())
		if out == "" {
			continue
		}
		// Handle current workspace denoted with a "*" prefix
		var current bool
		if strings.HasPrefix(out, "*") {
			var found bool
			_, out, found = strings.Cut(out, " ")
			if !found {
				return nil, fmt.Errorf("malformed output: %s", out)
			}
			current = true
		}
		ws := &Workspace{
			Resource: ID{Path: path, Name: out},
			Current:  current,
		}
		workspaces = append(workspaces, ws)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return workspaces, nil
}

// Create a workspace. Asynchronous.
func (s *Service) Create(path, name string) (*Workspace, *task.Task, error) {
	ws := &Workspace{Resource: ID{Path: path, Name: name}}

	// check if workspace exists already
	if _, err := s.Get(ws.Resource); err != nil {
		return nil, nil, resource.ErrExists
	}
	// check module exists
	if _, err := s.modules.Get(path); err != nil {
		return nil, nil, fmt.Errorf("checking for module: %s: %w", path, err)
	}
	task, err := s.modules.CreateTask(path, task.CreateOptions{
		Command: []string{"workspace", "new"},
		Args:    []string{name},
		AfterExited: func(*task.Task) {
			s.mu.Lock()
			s.workspaces[ws.Resource] = ws
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
func (s *Service) Get(id ID) (*Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id]
	if !ok {
		return nil, resource.ErrNotFound
	}
	return ws, nil
}

// List workspaces by module.
func (s *Service) ListByModule(path string) []*Workspace {
	s.mu.Lock()
	defer s.mu.Unlock()

	var workspaces []*Workspace
	for _, ws := range s.workspaces {
		if ws.ModulePath == path {
			workspaces = append(workspaces, ws)
		}
	}
	return workspaces
}

// Delete a workspace. Asynchronous.
func (s *Service) Delete(id ID) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id]
	if !ok {
		return nil, resource.ErrNotFound
	}

	return s.modules.CreateTask(id.Path, task.CreateOptions{
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
func (s *Service) CreateTask(id ID, opts task.CreateOptions) (*task.Task, error) {
	// ensure workspace exists
	ws, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %s: %w", id, err)
	}
	if opts.Metadata == nil {
		opts.Metadata = make(map[string]string)
	}
	opts.Metadata[MetadataKey] = id.Name
	opts.Env = append(opts.Env, fmt.Sprintf("TF_WORKSPACE=%s", id.Name))
	return s.modules.CreateTask(id.Path, opts)
}
