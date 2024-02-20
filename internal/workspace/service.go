package workspace

import (
	"bufio"
	"fmt"
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
func (s *Service) Reload() error {
	// TODO: once the workspace type has dynamic info, e.g. a status, then
	// consider not overwriting existing workspaces.
	s.mu.Lock()
	s.workspaces = make(map[ID]*Workspace)
	s.mu.Unlock()

	// standard log message
	msg := "reloading workspaces"

	for _, m := range s.modules.List() {
		mod := m
		if mod.Status != module.Initialized {
			// TODO: log message
			continue
		}
		tsk, err := s.tasks.Create(task.Spec{
			Args: []string{"workspace", "list"},
			Path: mod.Path,
		})
		if err != nil {
			slog.Error(msg, "error", err, "module", mod)
		}
		// TODO: pass in context
		go func() {
			if err := tsk.Wait(); err != nil {
				slog.Error(msg, "error", err, "module", mod)
				return
			}
			// should output something like this:
			//
			// 1> terraform workspace list
			//   default
			//   non-default-1
			// * non-default-2
			scanner := bufio.NewScanner(tsk.NewReader())
			for scanner.Scan() {
				out := strings.TrimSpace(scanner.Text())
				if out == "" {
					continue
				}
				if strings.HasPrefix(out, "*") {
					var found bool
					_, out, found = strings.Cut(out, " ")
					if !found {
						slog.Error(msg+": malformed output", "module", mod, "output", scanner.Text())
						continue
					}
				}
				ws := &Workspace{ID: ID{Path: mod.Path, Name: out}}

				s.mu.Lock()
				s.workspaces[ws.ID] = ws
				s.mu.Unlock()
			}
			if err := scanner.Err(); err != nil {
				slog.Error(msg, "error", err, "module", mod)
			}
		}()
	}
	return nil
}

// Create a workspace.
func (s *Service) Create(path, name string) (*Workspace, *task.Task, error) {
	ws := &Workspace{ID: ID{Path: path, Name: name}}

	// check if workspace exists already
	if _, err := s.Get(ws.ID); err != nil {
		return nil, nil, resource.ErrExists
	}
	// check module exists
	if _, err := s.modules.Get(path); err != nil {
		return nil, nil, fmt.Errorf("checking for module: %s: %w", path, err)
	}
	tsk, err := s.tasks.Create(task.Spec{
		Args: []string{"workspace", "new", name},
		Path: path,
	})
	if err != nil {
		return nil, nil, err
	}
	go func() {
		if err := tsk.Wait(); err != nil {
			// TODO: log error
			return
		}

		s.mu.Lock()
		s.workspaces[ws.ID] = ws
		s.mu.Unlock()

		s.broker.Publish(resource.CreatedEvent, ws)
	}()
	return ws, nil, nil
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
func (s *Service) ListByModule(path string) ([]*Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var workspaces []*Workspace
	for _, ws := range s.workspaces {
		if ws.Path == path {
			workspaces = append(workspaces, ws)
		}
	}
	return workspaces, nil
}

// Delete a workspace.
func (s *Service) Delete(id ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ws, ok := s.workspaces[id]
	if !ok {
		return resource.ErrNotFound
	}
	s.broker.Publish(resource.DeletedEvent, ws)
	delete(s.workspaces, id)
	return nil
}
