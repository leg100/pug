package state

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Service struct {
	tasks      *task.Service
	workspaces *workspace.Service
	modules    *module.Service
}

// Get asynchronously retrieves the state for a workspace.
func (s *Service) Get(workspaceID uuid.UUID) (*task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("getting state: %w", err)
	}
	mod, err := s.modules.Get(ws.Module().ID)
	if err != nil {
		return nil, fmt.Errorf("getting state: %w", err)
	}
	return s.tasks.Create(task.CreateOptions{
		Parent:  ws.Resource,
		Command: []string{"state", "pull"},
		Path:    mod.Path,
		Env:     []string{ws.TerraformEnv()},
	})
}

// ListResources retrieves the resources for a workspace.
// func (s *Service) ListResources(workspaceID uuid.UUID) ([]Resource, error) {
// 	get, err := s.Get(workspaceID)
// 	if err != nil {
// 		return nil, fmt.Errorf("listing resources: %w", err)
// 	}
// 	if err := get.Wait(); err != nil {
// 		return nil, err
// 	}
// 	ws, err := s.workspaces.Get(workspaceID)
// 	if err != nil {
// 		return nil, fmt.Errorf("getting state: %w", err)
// 	}
// 	mod, err := s.modules.Get(ws.Module().ID)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("creating run: %w", err)
// 	}
// 	// TODO: make the above a synchronous op
// 	var file File
// 	if err := json.NewDecoder(ws.NewReader()).Decode(&file); err != nil {
// 		return nil, err
// 	}
// 	return file.Resources, nil
// }

// RemoveItems removes items from the state. Aynchronous.
func (s *Service) RemoveItems(workspaceID uuid.UUID, addrs ...string) (*task.Task, error) {
	// create task invoking "terraform state rm [<addr>...]"
	//
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %s: %w", workspaceID, err)
	}
	mod, err := s.modules.Get(ws.Module().ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving module: %s: %w", ws.Module().ID, err)
	}
	return s.tasks.Create(task.CreateOptions{
		Parent:   ws.Resource,
		Blocking: true,
		Command:  []string{"state", "rm"},
		Args:     addrs,
		Path:     mod.Path,
		Env:      []string{ws.TerraformEnv()},
	})
}

func (s *Service) Taint(workspaceID uuid.UUID, addr string) (*task.Task, error) {
	ws, err := s.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %s: %w", workspaceID, err)
	}
	mod, err := s.modules.Get(ws.Module().ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving module: %s: %w", ws.Module().ID, err)
	}
	return s.tasks.Create(task.CreateOptions{
		Parent:   ws.Resource,
		Blocking: true,
		Command:  []string{"taint"},
		Args:     []string{addr},
		Path:     mod.Path,
		Env:      []string{ws.TerraformEnv()},
	})
}
